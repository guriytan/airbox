package controller

import (
	"net/http"
	"sync"

	"airbox/global"
	"airbox/logger"
	"airbox/model/vo"
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

var (
	info     *InfoController
	infoOnce sync.Once
)

func GetInfoController() *InfoController {
	infoOnce.Do(func() {
		info = &InfoController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			user:           service.GetUserService(),
		}
	})
	return info
}

// ListFile 显示文件和文件夹列表
func (i *InfoController) ListFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ListFile")
	req := vo.FileModel{}
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	fatherID := global.DefaultFatherID
	if req.FatherID != 0 {
		fatherID = req.FatherID
	}
	files, count, err := i.file.SelectFileByFatherID(ctx, i.GetAuth(c).Storage.ID, fatherID, req.Cursor, req.Limit)
	if err != nil {
		log.WithError(err).Warnf("get file by father_id failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"files": files, "total": count})
}

// UserInfo 显示用户及相关信息
func (i *InfoController) UserInfo(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "UserInfo")

	userInfo, err := i.user.GetUserByID(ctx, i.GetAuth(c).ID)
	if err != nil {
		log.WithError(err).Warnf("get user by id failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	count, err := i.file.SelectFileTypeCount(ctx, userInfo.Storage.ID)
	if err != nil {
		log.WithError(err).Warnf("get file statistics failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"user_info": userInfo, "count": count})
}

// ListType 显示对应类型的文件
func (i *InfoController) ListType(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ListType")
	req := vo.TypeModel{}
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	files, count, err := i.file.GetFileByType(ctx, i.GetAuth(c).Storage.ID, req.GetFatherID(), req.Type, req.Cursor, req.Limit)
	if err != nil {
		log.WithError(err).Warnf("get file by type failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"files": files, "total": count})
}

// ShareFile 分享文件
func (i *InfoController) ShareFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ShareFile")
	req := vo.ShareModel{}
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Link) == 0 {
		c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
		return
	}
	fileID, exp, err := encryption.ParseShareToken(req.Link)
	if err != nil {
		log.WithError(err).Warnf("parse token failed")
		c.JSON(http.StatusForbidden, global.ErrorOfWrongToken)
		return
	} else if exp < utils.Epoch() {
		c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
		return
	}
	fileByID, err := i.file.GetFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Warnf("get file by id failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	i.DownloadFile(c, fileByID)
}
