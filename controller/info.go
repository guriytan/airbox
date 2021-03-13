package controller

import (
	"net/http"
	"strconv"

	"airbox/global"
	"airbox/logger"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"

	"github.com/gin-gonic/gin"
)

type InfoController struct {
	BaseController
	file *service.FileService
	user *service.UserService
}

var info *InfoController

func GetInfoController() *InfoController {
	if info == nil {
		info = &InfoController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			user:           service.GetUserService(),
		}
	}
	return info
}

// ListFile 显示文件和文件夹列表
func (info *InfoController) ListFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ListFile")
	data := make(map[string]interface{})
	if fid := c.Query("fid"); fid != "" {
		files, err := info.file.SelectFileByFatherID(ctx, fid)
		if err != nil {
			log.Infof("%+v\n", err)
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
		data["files"] = files
	} else {
		sid := info.auth(c).Storage.ID
		files, err := info.file.GetFileByStorageID(ctx, sid)
		if err != nil {
			log.Infof("%+v\n", err)
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
		data["files"] = files
	}
	c.JSON(http.StatusOK, data)
}

// UserInfo 显示用户及相关信息
func (info *InfoController) UserInfo(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "UserInfo")
	data := make(map[string]interface{})
	user, err := info.user.GetUserByID(ctx, info.auth(c).ID)
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	count, err := info.file.SelectFileTypeCount(ctx, user.Storage.ID)
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	data["info"] = user
	data["count"] = count
	c.JSON(http.StatusOK, data)
}

// ListType 显示对应类型的文件
func (info *InfoController) ListType(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ListType")
	query := c.Query("type")
	fileType, _ := strconv.ParseInt(query, 10, 64)
	files, err := info.file.GetFileByType(ctx, int(fileType))
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"files": files,
	})
}

// ShareFile 分享文件
func (info *InfoController) ShareFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ShareFile")
	token := c.PostForm("link")
	if token == "" {
		c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
		return
	}
	fileID, exp, err := encryption.ParseShareToken(token)
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusForbidden, global.ErrorOfWrongToken)
		return
	} else if exp < utils.Epoch() {
		c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
		return
	}
	fileByID, err := info.file.GetFileByID(ctx, fileID)
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	info.downloadFile(c, fileByID)
}
