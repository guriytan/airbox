package controller

import (
	"airbox/config"
	"airbox/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strconv"
)

type FileController struct {
	BaseController
	file   *service.FileService
	folder *service.FolderService
	user   *service.UserService
}

var file *FileController

func GetFileController() *FileController {
	if file == nil {
		file = &FileController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			folder:         service.GetFolderService(),
			user:           service.GetUserService(),
		}
	}
	return file
}

// UploadFile 文件上传
func (f *FileController) UploadFile(c echo.Context) error {
	data := make(map[string]interface{})
	reader, err := c.Request().MultipartReader()
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	user, err := f.user.GetUserByID(f.auth(c).ID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	size, md5, sid, fid := uint64(0), "", user.Storage.ID, c.QueryParams().Get("fid")
	// 判断fid对应的文件夹是否存在
	if fid != "" {
		if _, err := f.folder.GetFolderByID(fid); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusBadRequest, err.Error())
		}
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
		switch part.FormName() {
		case "size":
			// 读取文件的大小
			s, err := readString(part)
			if err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusBadRequest, err.Error())
			}
			size, err = strconv.ParseUint(s, 10, 64)
			if err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusBadRequest, err.Error())
			}
			// 判断仓库的剩余容量是否足以存放该文件
			if user.Storage.CurrentSize+size >= user.Storage.MaxSize {
				_ = part.Close()
				return c.JSON(http.StatusBadRequest, config.ErrorOutOfSpace)
			}
		case "md5":
			// 读取文件的MD5
			md5, err = readString(part)
			if err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusBadRequest, err.Error())
			}
		case "folder":
			// 若文件存在前置文件夹，则需要调用folder service进行对文件夹进行分割
			// 并在数据库中创建每一层文件夹，最终返回最终层文件夹fid
			filepath, err := readString(part)
			if err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusBadRequest, err.Error())
			}
			if fid, err = f.folder.CreateFolder(filepath, sid, fid); err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		default:
			// 计算文件实际存储路径
			filepath := config.Env.Upload.Dir + "/" + sid + "/" + uuid.New().String() + "/"
			// 调用service方法保存文件数据
			filename, err := f.file.UploadFile(filepath, sid, fid, part, size)
			if err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
			file, err := f.file.StoreFile(md5, filename, filepath, size, sid, fid)
			if err != nil {
				_ = part.Close()
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
			data["file"] = file
		}
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

func readString(r io.Reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	return string(buf[:n]), nil
}
