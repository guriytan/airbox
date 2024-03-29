package controller

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/model/vo"
	"airbox/service"
	"airbox/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FileController struct {
	BaseController
	file *service.FileService
	user *service.UserService
}

var (
	file     *FileController
	fileOnce sync.Once
)

func GetFileController() *FileController {
	fileOnce.Do(func() {
		file = &FileController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			user:           service.GetUserService(),
		}
	})
	return file
}

// NewFile 新建空文件
func (f *FileController) NewFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "UploadFile")
	req := &vo.FileModel{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	userInfo, err := f.user.GetUserByID(ctx, f.GetAuth(c).ID)
	if err != nil {
		log.WithError(err).Warnf("get user failed")
		c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
		return
	}
	uploadFile, err := f.file.NewFile(ctx, userInfo.Storage.ID, req.FatherID, req.Name)
	if err != nil {
		log.WithError(err).Warnf("upload file failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"file": uploadFile})
}

// UploadFile 文件上传
func (f *FileController) UploadFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "UploadFile")
	req := &vo.FileModel{}
	if err := c.BindQuery(req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	userInfo, err := f.user.GetUserByID(ctx, f.GetAuth(c).ID)
	if err != nil {
		log.WithError(err).Warnf("get user failed")
		c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
		return
	}
	// 判断仓库的剩余容量是否足以存放该文件
	if userInfo.Storage.CurrentSize+req.Size >= userInfo.Storage.MaxSize {
		c.JSON(http.StatusBadRequest, global.ErrorOutOfSpace)
		return
	}
	reader, err := c.Request.MultipartReader()
	if err != nil {
		log.WithError(err).Warnf("get miltipart failed")
		c.JSON(http.StatusBadRequest, global.ErrorOfRequestParameter)
		return
	}
	uploadFile, err := f.uploadFile(ctx, reader, req, userInfo)
	if err != nil {
		log.WithError(err).Warnf("upload file failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"file": uploadFile})
}

func (f *FileController) uploadFile(ctx context.Context, reader *multipart.Reader, req *vo.FileModel, userInfo *do.User) (*do.File, error) {
	log := logger.GetLogger(ctx, "uploadFile")
	// 判断fid对应的文件夹是否存在
	// 获得Reader流
	part, err := reader.NextPart()
	// 若读取到结尾则跳出
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		log.WithError(err).Warnf("read multipart failed")
		return nil, err
	}
	defer func() { _ = part.Close() }()
	// 查找是否存在md5相同的文件
	fileByHash, err := f.file.SelectFileByHash(ctx, req.Hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 调用service方法保存文件数据
		fileByHash, err = f.file.UploadFile(ctx, &userInfo.Storage, part, req.Hash, req.Size)
		if err != nil {
			log.WithError(err).Warnf("upload file failed")
			return nil, err
		}
	} else if err != nil {
		log.WithError(err).Warnf("get file by hash failed")
		return nil, err
	}
	fileInfo, err := f.file.StoreFile(ctx, fileByHash, userInfo.Storage.ID, req.FatherID, part.FileName())
	if err != nil {
		log.WithError(err).Warnf("store file failed")
		return nil, err
	}
	return fileInfo, nil
}

// DownloadFile 文件下载
func (f *FileController) DownloadFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "DownloadFile")
	req := vo.FileModel{}
	if err := c.BindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	// 获取所要下载的文件信息
	fileByID, err := f.file.GetFileByID(ctx, req.FileID)
	if err != nil {
		log.WithError(err).Warnf("get file by id failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	f.BaseController.DownloadFile(c, fileByID)
}

// DeleteFile 删除文件
func (f *FileController) DeleteFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "Delete")
	req := vo.FileModel{}
	if err := c.BindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := f.file.DeleteFile(ctx, req.FileID); err != nil {
		log.WithError(err).Warnf("delete file failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.Status(http.StatusOK)
}

// UpdateFile 更新文件信息，包括重命名、移动和复制
func (f *FileController) UpdateFile(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "Update")
	req := vo.UpdateFileModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if req.FatherID == req.FileID {
		c.JSON(http.StatusBadRequest, global.ErrorOfCopyFile)
		return
	}
	switch req.OptType {
	case global.OperationTypeRename:
		// 重命名
		if err := f.file.RenameFile(ctx, req.FileID, req.Name); err != nil {
			log.WithError(err).Warnf("rename failed")
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
	case global.OperationTypeCopy:
		// 复制
		if err := f.file.CopyFile(ctx, req.FatherID, req.FileID); err != nil {
			log.WithError(err).Warnf("copy failed")
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
	case global.OperationTypeMove:
		// 移动
		if err := f.file.MoveFile(ctx, req.FatherID, req.FileID); err != nil {
			log.WithError(err).Warnf("move failed")
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
	}
	c.Status(http.StatusOK)
}
