package controller

import (
	"io"
	"net/http"
	"strconv"

	"airbox/global"
	"airbox/logger"
	"airbox/service"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FileController struct {
	BaseController
	file *service.FileService
	user *service.UserService
}

var file *FileController

func GetFileController() *FileController {
	if file == nil {
		file = &FileController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			user:           service.GetUserService(),
		}
	}
	return file
}

// UploadFile 文件上传
func (f *FileController) UploadFile(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "UploadFile")
	data := make(map[string]interface{})
	reader, err := c.Request().MultipartReader()
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
	}
	user, err := f.user.GetUserByID(ctx, f.auth(c).ID)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
	}
	size, hash, sid, fid := uint64(0), "", user.Storage.ID, c.QueryParams().Get("fid")
	// 判断fid对应的文件夹是否存在
	for {
		// 获得Reader流
		part, err := reader.NextPart()
		// 若读取到结尾则跳出
		if err == io.EOF {
			break
		} else if err != nil {
			log.Infof("%+v\n", err)
			return c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
		}
		switch part.FormName() {
		case "size":
			// 读取文件的大小
			s, err := readString(part)
			if err != nil {
				_ = part.Close()
				log.Infof("%+v\n", err)
				return c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
			}
			size, err = strconv.ParseUint(s, 10, 64)
			if err != nil {
				_ = part.Close()
				log.Infof("%+v\n", err)
				return c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
			}
			// 判断仓库的剩余容量是否足以存放该文件
			if user.Storage.CurrentSize+size >= user.Storage.MaxSize {
				_ = part.Close()
				return c.JSON(http.StatusBadRequest, global.ErrorOutOfSpace)
			}
		case "hash":
			// 读取文件的MD5
			hash, err = readString(part)
			if err != nil {
				_ = part.Close()
				log.Infof("%+v\n", err)
				return c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
			}
		case "folder":
			// 若文件存在前置文件夹，则需要调用folder service进行对文件夹进行分割
			// 并在数据库中创建每一层文件夹，最终返回最终层文件夹fid
		default:
			// 查找是否存在md5相同的文件
			fileByHash, err := f.file.SelectFileByHash(ctx, hash)
			if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && fileByHash.Size != size) {
				// 调用service方法保存文件数据
				fileByHash, err = f.file.UploadFile(ctx, part, sid, hash, size)
				if err != nil {
					_ = part.Close()
					log.Infof("%+v\n", err)
					return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
				}
			} else if err != nil {
				log.Infof("%+v\n", err)
				return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			}
			file, err := f.file.StoreFile(ctx, fileByHash, sid, fid)
			if err != nil {
				_ = part.Close()
				log.Infof("%+v\n", err)
				return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			}
			data["file"] = file
		}
		_ = part.Close()
	}
	return c.JSON(http.StatusOK, data)
}

// DownloadFile 文件下载
func (f *FileController) DownloadFile(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "DownloadFile")
	// 获取所要下载的文件信息
	fileByID, err := info.file.GetFileByID(ctx, c.Param("id"))
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
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

func readString(r io.Reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	return string(buf[:n]), nil
}
