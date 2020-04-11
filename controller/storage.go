package controller

import (
	"airbox/config"
	"airbox/service"
	"airbox/utils"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// StorageController is responsible for the file and folder operation
type StorageController struct {
	file   *service.FileService
	folder *service.FolderService
	*BaseController
}

var storage *StorageController

// GetStorageController return instance of StorageController
func GetStorageController() *StorageController {
	if storage == nil {
		storage = &StorageController{
			file:           service.GetFileService(),
			folder:         service.GetFolderService(),
			BaseController: GetBaseController(),
		}
	}
	return storage
}

// UploadFile 文件上传
func (s *StorageController) UploadFile(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	params := c.QueryParams()
	// fid为文件夹ID，size为文件大小
	fid, size := params.Get("fid"), params.Get("size")
	contentLength, err := strconv.ParseUint(size, 10, 64)
	if err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	reader, err := c.Request().MultipartReader()
	if err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	for {
		// 获得Reader流
		part, err := reader.NextPart()
		// 若读取到结尾则跳出
		if err == io.EOF {
			break
		} else if err != nil {
			data["warning"] = err.Error()
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusOK, data)
		}
		filename := part.FileName()
		if filename != "" {
			// 调用service方法保存文件数据
			err = s.file.UploadFile(part, filename, contentLength, s.Auth(c).ID, fid)
			if err != nil {
				data["warning"] = err.Error()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusOK, data)
			}
		}
		_ = part.Close()
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// DownloadFile 文件下载
func (s *StorageController) DownloadFile(c echo.Context) error {
	// 获取所要下载的文件信息
	id := c.QueryParam("id")
	if id == "" {
		http.Error(c.Response(), "缺少文件id", http.StatusBadRequest)
		return nil
	}
	return s.downloadFile(c, id)
}

// ShareFile 分享文件
func (s *StorageController) ShareFile(c echo.Context) error {
	token := c.FormValue("f")
	if token == "" {
		http.Error(c.Response(), "缺少token", http.StatusBadRequest)
		return nil
	}
	fileID, exp, err := utils.ParseShareToken(token)
	if err != nil {
		c.Logger().Warnf("failed to parse token: a", err)
		http.Error(c.Response(), "token错误", http.StatusForbidden)
		return nil
	} else if exp < utils.Epoch() {
		http.Error(c.Response(), config.ErrorOutOfDated.Error(), http.StatusUnauthorized)
		return nil
	}
	return s.downloadFile(c, fileID)
}

// downloadFile 公共使用的下载文件模块
func (s *StorageController) downloadFile(c echo.Context, id string) error {
	fileByID, err := s.file.GetFileByID(id)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		return err
	}
	open, err := os.Open(fileByID.Location + fileByID.Name)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		return err
	}
	stat, err := open.Stat()
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		return err
	}
	c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(stat.Name()))
	http.ServeContent(c.Response(), c.Request(), stat.Name(), stat.ModTime(), open)
	return nil
}

// GetShareLink 获取分享文件的token
func (s *StorageController) GetShareLink(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	id := c.FormValue("id")
	if id == "" {
		data["warning"] = "缺少文件id"
		data["code"] = config.CodeErrorOfRequest
		return c.JSON(http.StatusOK, data)
	}
	fileByID, err := s.file.GetFileByID(id)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		data["warning"] = err.Error()
		return c.JSON(http.StatusOK, data)
	}
	token, err := utils.GenerateShareToken(fileByID.ID)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		data["warning"] = err.Error()
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	data["link"] = token
	return c.JSON(http.StatusOK, data)
}

// DeleteFile 删除文件
func (s *StorageController) DeleteFile(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	id := c.QueryParam("id")
	if id == "" {
		data["code"] = config.CodeErrorOfRequest
		data["warning"] = "缺少文件id"
		return c.JSON(http.StatusOK, data)
	}
	if err := s.file.DeleteFile(id); err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// RenameFile 重命名文件
func (s *StorageController) RenameFile(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	name, id := c.FormValue("name"), c.FormValue("id")
	if name == "" || id == "" {
		data["code"] = config.CodeErrorOfRequest
		data["warning"] = "缺少信息"
		return c.JSON(http.StatusOK, data)
	}
	if err := s.file.Rename(name, id); err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// DeleteFolder 删除文件夹
func (s *StorageController) DeleteFolder(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	id := c.QueryParam("id")
	if id == "" {
		data["code"] = config.CodeErrorOfRequest
		data["warning"] = "缺少ID"
		return c.JSON(http.StatusOK, data)
	}
	if err := s.folder.DeleteFolder(id); err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// AddFolder 新建文件夹
func (s *StorageController) AddFolder(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	name, fid := c.FormValue("name"), c.FormValue("fid")
	if name == "" {
		data["code"] = config.CodeErrorOfRequest
		data["warning"] = "缺少名字"
		return c.JSON(http.StatusOK, data)
	}
	if err := s.folder.AddFolder(name, s.Auth(c).Storage.ID, fid); err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// RenameFolder 重命名文件夹
func (s *StorageController) RenameFolder(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfServer,
	}
	name, id := c.FormValue("name"), c.FormValue("id")
	if name == "" || id == "" {
		data["code"] = config.CodeErrorOfRequest
		data["warning"] = "缺少信息"
		return c.JSON(http.StatusOK, data)
	}
	if err := s.folder.Rename(name, id, s.Auth(c).Storage.ID, c.FormValue("fid")); err != nil {
		data["warning"] = err.Error()
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}
