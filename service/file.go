package service

import (
	"context"
	"mime/multipart"
	"path"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"

	"airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/utils"
	"airbox/utils/hasher"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FileService struct {
	file    base.FileDao
	info    base.FileInfoDao
	storage base.StorageDao
}

var (
	fileService     *FileService
	fileServiceOnce sync.Once
)

func GetFileService() *FileService {
	fileServiceOnce.Do(func() {
		fileService = &FileService{
			info:    db.GetFileInfoDao(pkg.GetDB()),
			file:    db.GetFileDao(pkg.GetDB()),
			storage: db.GetStorageDao(pkg.GetDB()),
		}
	})
	return fileService
}

// GetFileByID 获取文件信息，用于下载
func (f *FileService) GetFileByID(ctx context.Context, id string) (*do.File, error) {
	log := logger.GetLogger(ctx, "GetFileByID")
	fileByID, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		log.WithError(err).Infof("get file by id: %v failed", id)
		return nil, err
	}
	return fileByID, nil
}

// GetFileByStorageID 获取在仓库Sid下的文件，一般用于初始显示
func (f *FileService) GetFileByStorageID(ctx context.Context, sid string) ([]*do.File, error) {
	log := logger.GetLogger(ctx, "GetFileByStorageID")
	byStorageID, err := f.file.SelectFileByStorageID(ctx, sid)
	if err != nil {
		log.WithError(err).Infof("get file by storage id: %v failed", sid)
		return nil, err
	}
	return byStorageID, nil
}

// SelectFileByFatherID 获取在父节点fid下的文件
func (f *FileService) SelectFileByFatherID(ctx context.Context, fid string) (files []*do.File, err error) {
	log := logger.GetLogger(ctx, "SelectFileByFatherID")
	byFolderID, err := f.file.SelectFileByFatherID(ctx, fid)
	if err != nil {
		log.WithError(err).Infof("get file by storage id: %v failed", fid)
		return nil, err
	}
	return byFolderID, nil
}

// GetFileByType 获取类型为fileType的文件
func (f *FileService) GetFileByType(ctx context.Context, fileType int) ([]*do.File, error) {
	log := logger.GetLogger(ctx, "GetFileByType")
	byType, err := f.file.SelectFileByType(ctx, fileType)
	if err != nil {
		log.WithError(err).Infof("get file by type: %v failed", fileType)
		return nil, err
	}
	return byType, nil
}

// SelectFileTypeCount 获取不同类型文件的统计数量
func (f *FileService) SelectFileTypeCount(ctx context.Context, sid string) (types []*do.Statistics, err error) {
	log := logger.GetLogger(ctx, "SelectFileTypeCount")
	typeCount, err := f.file.SelectFileTypeCount(ctx, sid)
	if err != nil {
		log.WithError(err).Infof("get file count by storage: %v failed", sid)
		return nil, err
	}
	return typeCount, nil
}

// UploadFile 保存文件信息，并更新数据仓库的容量大小
func (f *FileService) UploadFile(ctx context.Context, storage *do.Storage, part *multipart.Part, hash string, size int64) (*do.FileInfo, error) {
	log := logger.GetLogger(ctx, "UploadFile")
	ossKey := hasher.GetMD5().Hash(storage.BucketName, part.FileName(), hash)
	fileInfo := &do.FileInfo{
		Hash:   hash,
		Name:   part.FileName(),
		OssKey: ossKey,
		Size:   size,
	}
	err := pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.info.InsertFileInfo(ctx, fileInfo); err != nil {
			log.WithError(err).Warnf("save file info: %+v failed", fileInfo)
			return err
		}
		info, err := pkg.GetOSS().PutObject(ctx, storage.BucketName, ossKey, part, size, minio.PutObjectOptions{})
		if err != nil {
			log.WithError(err).Warnf("put object: %v size: %v to oss: %v failed", ossKey, size, storage.BucketName)
			return err
		}
		log.Infof("put file: %+v to oss: %+v success", fileInfo, info)
		return nil
	})
	if err != nil {
		log.WithError(err).Warnf("transaction failed")
		return nil, err
	}
	return fileInfo, nil
}

// StoreFile 保存文件信息并更新仓库现容量大小
func (f *FileService) StoreFile(ctx context.Context, info *do.FileInfo, sid, fid string) (*do.File, error) {
	log := logger.GetLogger(ctx, "StoreFile")
	filename := info.Name
	// 判断是否已存在同名文件并修改文件名（增加数字编号）
	if _, err := f.file.SelectFileByName(ctx, filename, sid, fid); err != nil {
		return nil, err
	} else {
		filename = utils.AddSuffixToFilename(filename)
	}
	file := &do.File{
		Name:      filename,
		StorageID: sid,
		FatherID:  global.DefaultFatherID,
		Type:      int(utils.GetFileType(strings.ToLower(path.Ext(filename)))),
		FileInfo:  *info,
	}
	if len(fid) != 0 {
		file.FatherID = fid
	}
	err := pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.InsertFile(ctx, tx, file); err != nil {
			log.WithError(err).Warnf("save file: %+v failed", file)
			return err
		}
		if err := f.storage.UpdateCurrentSize(ctx, tx, sid, info.Size); err != nil {
			log.WithError(err).Warnf("update storage: %v, size: %v failed", sid, info.Size)
			return err
		}
		if err := f.info.UpdateFileInfo(ctx, tx, info.ID, 1); err != nil {
			log.WithError(err).Warnf("update file info: %v failed", info.ID)
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Warn("transaction failed")
		return nil, err
	}
	return file, nil
}

// DeleteFile 删除文件信息，并更新数据仓库的容量大小
func (f *FileService) DeleteFile(ctx context.Context, storage *do.Storage, id string) error {
	log := logger.GetLogger(ctx, "DeleteFile")
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		return err
	}
	err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.DeleteFileByID(ctx, id); err != nil {
			log.WithError(err).Warnf("delete file: %v failed", id)
			return err
		}
		if err := f.storage.UpdateCurrentSize(ctx, tx, file.StorageID, -file.FileInfo.Size); err != nil {
			log.WithError(err).Warnf("update storage: %v size: %v failed", file.StorageID, -file.FileInfo.Size)
			return err
		}
		if file.FileInfo.Link > 1 {
			if err := f.info.UpdateFileInfo(ctx, tx, file.FileInfoID, -1); err != nil {
				log.WithError(err).Warnf("update file info: %v failed", file.FileInfoID)
				return err
			}
		} else {
			if err := f.info.DeleteFileInfo(ctx, tx, file.FileInfoID); err != nil {
				log.WithError(err).Warnf("delete file info: %v failed", file.FileInfoID)
				return err
			}
			if err := pkg.GetOSS().RemoveObject(ctx, storage.BucketName, file.FileInfo.OssKey, minio.RemoveObjectOptions{}); err != nil {
				log.WithError(err).Warnf("remove object: %v from bucket: %v failed", file.FileInfo.OssKey, storage.BucketName)
				return err
			}
			log.Infof("remove file: %v from bucket: %v success", file.FileInfo.OssKey, storage.BucketName)
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Warn("transaction failed")
		return err
	}
	log.Infof("delete file: %v success", id)
	return nil
}

// RenameFile 重命名，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) RenameFile(ctx context.Context, id, name string) error {
	log := logger.GetLogger(ctx, "RenameFile")
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", id)
		return err
	}
	exist, err := f.file.SelectFileByName(ctx, name, file.StorageID, file.FatherID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Warnf("get file by name: %v in storage: %v, father: %v failed", name, file.StorageID, file.FatherID)
		return err
	} else if exist != nil {
		return errors.New(global.ErrorOfConflict)
	}
	if err := f.file.UpdateFile(ctx, id, map[string]interface{}{"name": name}); err != nil {
		log.WithError(err).Warnf("update file: %v name: %v failed", id, name)
		return err
	}
	log.Infof("update file: %v name: %v success", id, name)
	return nil
}

// MoveFile 移动文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) MoveFile(ctx context.Context, fid, id string) error {
	log := logger.GetLogger(ctx, "MoveFile")
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", id)
		return err
	}
	exist, err := f.file.SelectFileByName(ctx, file.Name, file.StorageID, fid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Warnf("get file by name: %v in storage: %v, father: %v failed", file.Name, file.StorageID, fid)
		return err
	} else if exist != nil {
		return errors.New(global.ErrorOfConflict)
	}
	save := map[string]interface{}{"father_id": global.DefaultFatherID}
	if len(fid) != 0 {
		save["father_id"] = fid
	}
	if err = f.file.UpdateFile(ctx, id, save); err != nil {
		log.WithError(err).Warnf("update file: %v father: %v failed", id, fid)
		return err
	}
	log.Infof("update file: %v father: %v success", id, fid)
	return nil
}

// CopyFile 复制文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) CopyFile(ctx context.Context, id, fid string) error {
	log := logger.GetLogger(ctx, "CopyFile")
	file, err := f.file.SelectFileByID(ctx, id)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", id)
		return err
	}
	_, err = f.StoreFile(ctx, &file.FileInfo, file.StorageID, fid)
	if err != nil {
		log.WithError(err).Warnf("store file: %v to father: %v of info: %+v failed", id, fid, file.FileInfo)
		return err
	}
	log.Infof("store file: %v to father: %v of info: %+v success", id, fid, file.FileInfo)
	return nil
}

// SelectFileByHash 根据文件Hash查询文件信息
func (f *FileService) SelectFileByHash(ctx context.Context, hash string) (*do.FileInfo, error) {
	log := logger.GetLogger(ctx, "SelectFileByHash")
	byMD5, err := f.info.SelectFileInfoByHash(ctx, hash)
	if err != nil {
		log.WithError(err).Warnf("get file info by hash: %v failed", hash)
		return nil, err
	}
	return byMD5, nil
}
