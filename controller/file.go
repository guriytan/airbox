package controller

import (
	"airbox/service"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strconv"
)

type FileController struct {
	BaseController
	file *service.FileService
}

var file *FileController

func GetFileController() *FileController {
	if file == nil {
		file = &FileController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
		}
	}
	return file
}

// UploadFile 文件上传
func (f *FileController) UploadFile(c echo.Context) error {
	data := make(map[string]interface{})
	params := c.QueryParams()
	// fid为文件夹ID，size为文件大小
	fid, size := params.Get("fid"), params.Get("size")
	contentLength, err := strconv.ParseUint(size, 10, 64)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	reader, err := c.Request().MultipartReader()
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	for {
		// 获得Reader流
		part, err := reader.NextPart()
		// 若读取到结尾则跳出
		if err == io.EOF {
			break
		} else if err != nil {
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		// 调用service方法保存文件数据
		file, err := f.file.UploadFile(part, contentLength, f.auth(c).ID, fid)
		if err != nil {
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		data["file"] = file
		_ = part.Close()
	}
	return c.JSON(http.StatusOK, data)
}

// DownloadFile 文件下载
func (f *FileController) DownloadFile(c echo.Context) error {
	// 获取所要下载的文件信息
	fileByID, err := info.file.GetFileByID(c.Param("id"))
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return f.downloadFile(c, fileByID)
}

// DeleteFile 删除文件
func (f *FileController) DeleteFile(c echo.Context) error {
	return f.Delete(c, f.file.DeleteFile)
}

// UpdateFile 更新文件信息，包括重命名、移动和复制
func (f *FileController) UpdateFile(c echo.Context) error {
	return f.Update(c, f.file.RenameFile, f.file.CopyFile, f.file.MoveFile)
}
