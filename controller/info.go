package controller

import (
	"net/http"

	"airbox/global"
	"airbox/logger"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"

	"github.com/labstack/echo/v4"
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
func (info *InfoController) ListFile(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "ListFile")
	data := make(map[string]interface{})
	if fid := c.QueryParam("fid"); fid != "" {
		files, err := info.file.SelectFileByFatherID(ctx, fid)
		if err != nil {
			log.Infof("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["files"] = files
	} else {
		sid := info.auth(c).Storage.ID
		files, err := info.file.GetFileByStorageID(ctx, sid)
		if err != nil {
			log.Infof("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["files"] = files
	}
	return c.JSON(http.StatusOK, data)
}

// UserInfo 显示用户及相关信息
func (info *InfoController) UserInfo(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "UserInfo")
	data := make(map[string]interface{})
	user, err := info.user.GetUserByID(ctx, info.auth(c).ID)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	count, err := info.file.SelectFileTypeCount(ctx, user.Storage.ID)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	data["info"] = user
	data["count"] = count
	return c.JSON(http.StatusOK, data)
}

// ListType 显示对应类型的文件
func (info *InfoController) ListType(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "ListType")
	files, err := info.file.GetFileByType(ctx, c.QueryParam("type"))
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"files": files,
	})
}

// ShareFile 分享文件
func (info *InfoController) ShareFile(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "ShareFile")
	token := c.FormValue("link")
	if token == "" {
		return c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
	}
	fileID, exp, err := encryption.ParseShareToken(token)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusForbidden, global.ErrorOfWrongToken)
	} else if exp < utils.Epoch() {
		return c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
	}
	fileByID, err := info.file.GetFileByID(ctx, fileID)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return info.downloadFile(c, fileByID)
}
