package controller

import (
	"net/http"
	"net/url"

	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/pkg"
	"airbox/utils"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// BaseController is responsible for the common operation of controller
type BaseController struct {
}

// getBaseController return instance of BaseController
func getBaseController() BaseController {
	return BaseController{}
}

// GetAuth verify return the authority of request
func (*BaseController) GetAuth(c *gin.Context) *do.User {
	userInfo, ok := c.Get(global.KeyAuthorization)
	if ok {
		return userInfo.(*do.User)
	}
	return nil
}

// DownloadFile 公共使用的下载文件模块
func (b *BaseController) DownloadFile(c *gin.Context, file *do.File) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "DownloadFile")
	storage := b.GetAuth(c).Storage
	object, err := pkg.GetOSS().GetObject(ctx, storage.BucketName, file.FileInfo.OssKey, minio.GetObjectOptions{})
	if err != nil {
		log.WithError(err).Warnf("get object: %v from bucket: %v failed", file.FileInfo.OssKey, storage.BucketName)
		c.JSON(http.StatusInternalServerError, global.ErrorDownloadFile)
		return
	}
	defer func() { _ = object.Close() }()
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(file.Name))
	http.ServeContent(c.Writer, c.Request, file.Name, file.UpdatedAt, object)
}
