package service

import (
	"context"
	"mime/multipart"
	"path"
	"strings"
	"sync"

	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/pkg"
	"airbox/utils"
	"airbox/utils/hasher"

	"github.com/minio/minio-go/v7"
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
func (f *FileService) GetFileByID(ctx context.Context, fileID int64) (*do.File, error) {
	log := logger.GetLogger(ctx, "GetFileByID")
	fileByID, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Infof("get file by id: %v failed", fileID)
		return nil, err
	}
	return fileByID, nil
}

// SelectFileByFatherID 获取在父节点fid下的文件
func (f *FileService) SelectFileByFatherID(ctx context.Context, fatherID, cursor int64, limit int) (files []*do.File, err error) {
	log := logger.GetLogger(ctx, "SelectFileByFatherID")
	folders, err := f.file.SelectFileByFatherIDAndType(ctx, fatherID, []int{int(global.FileFolderType)}, cursor, limit)
	if err != nil {
		log.WithError(err).Infof("get file by storage id: %v failed", fatherID)
		return nil, err
	}
	files = append(files, folders...)

	fileTypes := []int{int(global.FileOtherType), int(global.FileMusicType), int(global.FileVideoType), int(global.FileDocumentType), int(global.FilePictureType)}
	byFolderID, err := f.file.SelectFileByFatherIDAndType(ctx, fatherID, fileTypes, cursor, limit)
	if err != nil {
		log.WithError(err).Infof("get file by storage id: %v failed", fatherID)
		return nil, err
	}
	files = append(files, byFolderID...)
	return files, nil
}

// GetFileByType 获取类型为fileType的文件
func (f *FileService) GetFileByType(ctx context.Context, fatherID int64, fileType int, cursor int64, limit int) ([]*do.File, error) {
	log := logger.GetLogger(ctx, "GetFileByType")
	byType, err := f.file.SelectFileByFatherIDAndType(ctx, fatherID, []int{fileType}, cursor, limit)
	if err != nil {
		log.WithError(err).Infof("get file by type: %v failed", fileType)
		return nil, err
	}
	return byType, nil
}

// SelectFileTypeCount 获取不同类型文件的统计数量
func (f *FileService) SelectFileTypeCount(ctx context.Context, storageID int64) (types []*do.Statistics, err error) {
	log := logger.GetLogger(ctx, "SelectFileTypeCount")
	typeCount, err := f.file.SelectFileTypeCount(ctx, storageID)
	if err != nil {
		log.WithError(err).Infof("get file count by storage: %v failed", storageID)
		return nil, err
	}
	return typeCount, nil
}

// NewFile 新建文件
func (f *FileService) NewFile(ctx context.Context, storageID, fatherID int64, name string, fileType global.FileType) (file *do.File, err error) {
	log := logger.GetLogger(ctx, "StoreFile")
	filename, err := f.FixFilename(ctx, name, storageID, fatherID)
	if err != nil {
		log.WithError(err).Warnf("fix name: %+v failed", name)
		return nil, err
	}
	err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		fileInfo := &do.FileInfo{
			Hash: hasher.GetSha256().Hash(storageID, fatherID, filename),
			Name: filename,
		}
		if err := f.info.InsertFileInfo(ctx, tx, fileInfo); err != nil {
			log.WithError(err).Warnf("save file info: %+v failed", fileInfo)
			return err
		}
		file = &do.File{
			Name:       filename,
			StorageID:  storageID,
			FatherID:   global.DefaultFatherID,
			Type:       int(fileType),
			FileInfoID: fileInfo.ID,
		}
		if fatherID != 0 {
			file.FatherID = fatherID
		}
		if err := f.file.InsertFile(ctx, tx, file); err != nil {
			log.WithError(err).Warnf("save file: %+v failed", file)
			return err
		}
		return nil
	})
	return file, nil
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
		if err := f.info.InsertFileInfo(ctx, tx, fileInfo); err != nil {
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
func (f *FileService) StoreFile(ctx context.Context, info *do.FileInfo, storageID, fatherID int64) (*do.File, error) {
	log := logger.GetLogger(ctx, "StoreFile")
	filename, err := f.FixFilename(ctx, info.Name, storageID, fatherID)
	if err != nil {
		log.WithError(err).Warnf("fix name: %+v failed", info.Name)
		return nil, err
	}
	file := &do.File{
		Name:      filename,
		StorageID: storageID,
		FatherID:  global.DefaultFatherID,
		Type:      int(utils.GetFileType(strings.ToLower(path.Ext(filename)))),
		FileInfo:  *info,
	}
	if fatherID != 0 {
		file.FatherID = fatherID
	}
	err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.InsertFile(ctx, tx, file); err != nil {
			log.WithError(err).Warnf("save file: %+v failed", file)
			return err
		}
		if err := f.storage.UpdateCurrentSize(ctx, tx, storageID, info.Size); err != nil {
			log.WithError(err).Warnf("update storage: %v, size: %v failed", storageID, info.Size)
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
func (f *FileService) DeleteFile(ctx context.Context, storage *do.Storage, fileID int64) error {
	log := logger.GetLogger(ctx, "DeleteFile")
	file, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		return err
	}
	err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.DeleteFileByID(ctx, tx, fileID); err != nil {
			log.WithError(err).Warnf("delete file: %v failed", fileID)
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
	log.Infof("delete file: %v success", fileID)
	return nil
}

// RenameFile 重命名，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) RenameFile(ctx context.Context, fileID int64, name string) error {
	log := logger.GetLogger(ctx, "RenameFile")
	file, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", fileID)
		return err
	}
	exist, err := f.file.SelectFileByName(ctx, name, file.StorageID, file.FatherID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Warnf("get file by name: %v in storage: %v, father: %v failed", name, file.StorageID, file.FatherID)
		return err
	} else if len(exist) != 0 {
		return errors.New(global.ErrorOfConflict)
	}
	if err := f.file.UpdateFile(ctx, fileID, map[string]interface{}{"name": name}); err != nil {
		log.WithError(err).Warnf("update file: %v name: %v failed", fileID, name)
		return err
	}
	log.Infof("update file: %v name: %v success", fileID, name)
	return nil
}

// MoveFile 移动文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) MoveFile(ctx context.Context, fatherID, fileID int64) error {
	log := logger.GetLogger(ctx, "MoveFile")
	file, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", fileID)
		return err
	}
	exist, err := f.file.SelectFileByName(ctx, file.Name, file.StorageID, fatherID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Warnf("get file by name: %v in storage: %v, father: %v failed", file.Name, file.StorageID, fatherID)
		return err
	} else if len(exist) != 0 {
		return errors.New(global.ErrorOfConflict)
	}
	save := map[string]interface{}{"father_id": global.DefaultFatherID}
	if fatherID != 0 {
		save["father_id"] = fatherID
	}
	if err = f.file.UpdateFile(ctx, fileID, save); err != nil {
		log.WithError(err).Warnf("update file: %v father: %v failed", fileID, fatherID)
		return err
	}
	log.Infof("update file: %v father: %v success", fileID, fatherID)
	return nil
}

// CopyFile 复制文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) CopyFile(ctx context.Context, fileID, fatherID int64) error {
	log := logger.GetLogger(ctx, "CopyFile")
	file, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", fileID)
		return err
	}
	_, err = f.StoreFile(ctx, &file.FileInfo, file.StorageID, fatherID)
	if err != nil {
		log.WithError(err).Warnf("store file: %v to father: %v of info: %+v failed", fileID, fatherID, file.FileInfo)
		return err
	}
	log.Infof("store file: %v to father: %v of info: %+v success", fileID, fatherID, file.FileInfo)
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

// FixFilename 获取最终文件名
func (f *FileService) FixFilename(ctx context.Context, name string, storageID, fatherID int64) (string, error) {
	// 判断是否已存在同名文件并修改文件名（增加数字编号）
	if _, err := f.file.SelectFileByName(ctx, name, storageID, fatherID); errors.Is(err, gorm.ErrRecordNotFound) {
		return name, nil
	} else if err != nil {
		return "", err
	} else {
		name = utils.AddSuffixToFilename(name)
	}
	return name, nil
}
