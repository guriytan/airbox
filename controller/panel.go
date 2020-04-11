package controller

import (
	"airbox/config"
	"airbox/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

// PanelController is responsible for the file data
type PanelController struct {
	user   *service.UserService
	file   *service.FileService
	folder *service.FolderService
	*BaseController
}

var panel *PanelController

// GetPanelController return instance of PanelController
func GetPanelController() *PanelController {
	if panel == nil {
		panel = &PanelController{
			user:           service.GetUserService(),
			file:           service.GetFileService(),
			folder:         service.GetFolderService(),
			BaseController: GetBaseController(),
		}
	}
	return panel
}

// ListFile 显示文件和文件夹列表
func (p *PanelController) ListFile(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	if fid := c.QueryParam("fid"); fid != "" {
		folders, err := p.folder.GetFolderByFatherID(fid)
		if err != nil {
			data["warning"] = err.Error()
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusOK, data)
		}
		data["folders"] = folders
		files, err := p.file.SelectFileByFolderID(fid)
		if err != nil {
			data["warning"] = err.Error()
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusOK, data)
		}
		data["files"] = files
		path, err := p.folder.GetFolderByID(fid)
		if err != nil {
			data["warning"] = err.Error()
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusOK, data)
		}
		data["path"] = path
	} else {
		sid := p.Auth(c).Storage.ID
		folders, err := p.folder.GetFolderByStorageID(sid)
		if err != nil {
			data["warning"] = err.Error()
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusOK, data)
		}
		data["folders"] = folders
		files, err := p.file.GetFileByStorageID(sid)
		if err != nil {
			data["warning"] = err.Error()
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusOK, data)
		}
		data["files"] = files
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// UserInfo 显示用户及相关信息
func (p *PanelController) UserInfo(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	count, err := p.file.SelectFileTypeCount()
	if err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	userByID, err := p.user.GetUserByID(p.Auth(c).ID)
	if err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["info"] = userByID
	data["count"] = count
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// ListType 显示对应类型的文件
func (p *PanelController) ListType(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	t := c.QueryParam("type")
	files, err := p.file.GetFileByType(t)
	if err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	data["files"] = files
	return c.JSON(http.StatusOK, data)
}

// ApplyToUnsubscribe 申请注销账户，向用户邮箱发送验证码邮件验证权限
func (p *PanelController) ApplyToUnsubscribe(c echo.Context) error {
	user := p.Auth(c)
	// 发送验证码至邮箱
	go func() {
		if err := p.user.SendCaptcha(user.Email); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
	}()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": config.CodeSuccess,
	})
}

// Unsubscribe 注销账户
func (p *PanelController) Unsubscribe(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	user := p.Auth(c)
	code := c.FormValue("code")
	// 将email作为key从缓存中提取验证码比对
	if !p.user.VerifyEmailCaptcha(user.Email, code) {
		data["warning"] = "验证码错误"
		return c.JSON(http.StatusOK, data)
	}
	// 从数据库中删除相关信息并从磁盘删除文件
	if err := p.user.CloseUser(user.ID, user.Storage.ID); err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}
