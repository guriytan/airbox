package controller

import (
	"airbox/global"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
)

type InfoController struct {
	BaseController
	file   *service.FileService
	folder *service.FolderService
	user   *service.UserService
}

var info *InfoController

func GetInfoController() *InfoController {
	if info == nil {
		info = &InfoController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			folder:         service.GetFolderService(),
			user:           service.GetUserService(),
		}
	}
	return info
}

// ListFile 显示文件和文件夹列表
func (info *InfoController) ListFile(c echo.Context) error {
	data := make(map[string]interface{})
	if fid := c.QueryParam("fid"); fid != "" {
		folders, err := info.folder.GetFolderByFatherID(fid)
		if err != nil {
			global.LOGGER.Printf("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["folders"] = folders
		files, err := info.file.SelectFileByFolderID(fid)
		if err != nil {
			global.LOGGER.Printf("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["files"] = files
		path, err := info.folder.GetFolderByIDWithPath(fid)
		if err != nil {
			global.LOGGER.Printf("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["path"] = path
	} else {
		sid := info.auth(c).Storage.ID
		folders, err := info.folder.GetFolderByStorageID(sid)
		if err != nil {
			global.LOGGER.Printf("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["folders"] = folders
		files, err := info.file.GetFileByStorageID(sid)
		if err != nil {
			global.LOGGER.Printf("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		data["files"] = files
	}
	return c.JSON(http.StatusOK, data)
}

// UserInfo 显示用户及相关信息
func (info *InfoController) UserInfo(c echo.Context) error {
	data := make(map[string]interface{})
	user, err := info.user.GetUserByID(info.auth(c).ID)
	if err != nil {
		global.LOGGER.Printf("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	count, err := info.file.SelectFileTypeCount(user.Storage.ID)
	if err != nil {
		global.LOGGER.Printf("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	data["info"] = user
	data["count"] = count
	return c.JSON(http.StatusOK, data)
}

// ListType 显示对应类型的文件
func (info *InfoController) ListType(c echo.Context) error {
	files, err := info.file.GetFileByType(c.QueryParam("type"))
	if err != nil {
		global.LOGGER.Printf("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"files": files,
	})
}

// ShareFile 分享文件
func (info *InfoController) ShareFile(c echo.Context) error {
	token := c.FormValue("link")
	if token == "" {
		return c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
	}
	fileID, exp, err := encryption.ParseShareToken(token)
	if err != nil {
		global.LOGGER.Printf("failed to parse token: %+v\n", errors.WithStack(err))
		return c.JSON(http.StatusForbidden, global.ErrorOfWrongToken)
	} else if exp < utils.Epoch() {
		return c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
	}
	fileByID, err := info.file.GetFileByID(fileID)
	if err != nil {
		global.LOGGER.Printf("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return info.downloadFile(c, fileByID)
}
