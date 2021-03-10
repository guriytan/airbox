package service

import (
	"context"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"sync"

	"airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/logger"
	"airbox/model"
	"airbox/utils"
	"airbox/utils/disk"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FileService struct {
	file    base.FileDao
	info    base.FileInfoDao
	storage base.StorageDao
}

// GetFileByID 获取文件信息，用于下载
func (f *FileService) GetFileByID(ctx context.Context, id string) (*model.File, error) {
	fileByID, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return fileByID, nil
}

// GetFileByStorageID 获取在仓库Sid下的文件，一般用于初始显示
func (f *FileService) GetFileByStorageID(ctx context.Context, sid string) ([]*model.File, error) {
	byStorageID, err := f.file.SelectFileByStorageID(ctx, sid)
	if err != nil {
		return nil, err
	}
	return byStorageID, nil
}

// SelectFileByFatherID 获取在父文件夹fid下的文件
func (f *FileService) SelectFileByFatherID(ctx context.Context, fid string) (files []*model.File, err error) {
	byFolderID, err := f.file.SelectFileByFatherID(ctx, fid)
	if err != nil {
		return nil, err
	}
	return byFolderID, nil
}

// GetFileByType 获取类型为t的文件
func (f *FileService) GetFileByType(ctx context.Context, t string) ([]*model.File, error) {
	byType, err := f.file.SelectFileByType(ctx, t)
	if err != nil {
		return nil, err
	}
	return byType, nil
}

// SelectFileTypeCount 获取不同类型文件的统计数量
func (f *FileService) SelectFileTypeCount(ctx context.Context, sid string) (types []*model.Statistics, err error) {
	typeCount, err := f.file.SelectFileTypeCount(ctx, sid)
	if err != nil {
		return nil, err
	}
	return typeCount, nil
}

// UploadFile 保存文件信息，并更新数据仓库的容量大小
func (f *FileService) UploadFile(ctx context.Context, part *multipart.Part, sid, md5 string, size uint64) (*model.FileInfo, error) {
	// 计算文件实际存储路径
	filepath := config.GetConfig().Upload.Dir + "/" + sid + "/" + uuid.New().String() + "/"
	filename := part.FileName()
	// 由于在上传文件夹模型下Filename()将有前置文件夹路径。因此统一剪切
	filename = filename[strings.LastIndex(filename, "/")+1:]
	if filename == "" {
		return nil, errors.New("without file")
	}
	// 调用Upload上传并返回文件长度
	if err := disk.Upload(filepath, filename, part, size); err != nil {
		return nil, err
	}
	info := &model.FileInfo{
		Hash: md5,
		Name: filename,
		Size: size,
		Path: filepath,
	}
	err := f.info.InsertFileInfo(ctx, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// StoreFile 保存文件信息并更新仓库现容量大小
func (f *FileService) StoreFile(ctx context.Context, info *model.FileInfo, sid, fid string) (*model.File, error) {
	filename := info.Name
	// 判断是否已存在同名文件并修改文件名（增加数字编号）
	if _, err := f.file.SelectFileByName(ctx, filename, sid, fid); err != nil {
		return nil, err
	} else {
		filename = utils.AddSuffixToFilename(filename)
	}
	file := &model.File{
		Name:      filename,
		StorageID: sid,
		Type:      int(utils.GetFileType(strings.ToLower(path.Ext(filename)))),
		FileInfo:  *info,
	}
	if len(fid) != 0 {
		file.FatherID = fid
	}
	err := config.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.InsertFile(ctx, tx, file); err != nil {
			return err
		}
		if err := f.storage.UpdateCurrentSize(ctx, tx, sid, int64(info.Size)); err != nil {
			return err
		}
		if err := f.info.UpdateFileInfo(ctx, tx, info.ID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return file, nil
}

// DeleteFile 删除文件信息，并更新数据仓库的容量大小
func (f *FileService) DeleteFile(ctx context.Context, id string) error {
	log := logger.GetLogger(ctx, "DeleteFile")
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		return err
	}
	err = config.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.DeleteFileByID(ctx, id); err != nil {
			return err
		}
		if err := f.storage.UpdateCurrentSize(ctx, tx, file.StorageID, int64(-file.FileInfo.Size)); err != nil {
			return err
		}
		if file.FileInfo.Link > 1 {
			if err := f.info.UpdateFileInfo(ctx, tx, file.FileInfoID, -1); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if err := f.info.DeleteFileInfo(ctx, tx, file.FileInfoID); err != nil {
				tx.Rollback()
				return err
			}
			go func() {
				err = os.RemoveAll(file.FileInfo.Path)
				if err != nil {
					log.Infof("%s\n", err.Error())
				}
			}()
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// RenameFile 重命名，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) RenameFile(ctx context.Context, name, id string) error {
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		return err
	}
	var fid = global.DefaultFatherID // go 不支持三元表达式，因此需要额外处理
	if len(file.FatherID) != 0 {
		fid = file.FatherID
	}
	if _, err = f.file.SelectFileByName(ctx, name, file.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	if err := f.file.UpdateFile(ctx, id, map[string]interface{}{"name": name}); err != nil {
		return err
	}
	return nil
}

// MoveFile 移动文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) MoveFile(ctx context.Context, fid, id string) error {
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err = f.file.SelectFileByName(ctx, file.Name, file.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	}
	save := make(map[string]interface{})
	if fid != "" {
		save["father_id"] = fid
	} else {
		save["father_id"] = global.DefaultFatherID
	}
	return f.file.UpdateFile(ctx, id, save)
}

// CopyFile 复制文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) CopyFile(ctx context.Context, fid, id string) error {
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		return err
	}
	_, err = f.StoreFile(ctx, &file.FileInfo, file.StorageID, fid)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileService) SelectFileByHash(ctx context.Context, hash string) (*model.FileInfo, error) {
	byMD5, err := f.info.SelectFileInfoByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	return byMD5, nil
}

var (
	fileService     *FileService
	fileServiceOnce sync.Once
)

func GetFileService() *FileService {
	fileServiceOnce.Do(func() {
		fileService = &FileService{
			info:    db.GetFileInfoDao(config.GetDB()),
			file:    db.GetFileDao(config.GetDB()),
			storage: db.GetStorageDao(config.GetDB()),
		}
	})
	return fileService
}
