package service

import (
	"context"
	"mime/multipart"
	"path"
	"strings"
	"sync"
	"time"

	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/model/dto"
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
func (f *FileService) SelectFileByFatherID(ctx context.Context, storageID, fatherID, cursor int64, limit int) (files []*do.File, count int64, err error) {
	log := logger.GetLogger(ctx, "SelectFileByFatherID")
	cond := &dto.QueryCondition{StorageID: storageID, FatherID: &fatherID, Cursor: cursor, Limit: limit}
	byFolderID, err := f.file.SelectFileByCondition(ctx, cond)
	if err != nil {
		log.WithError(err).Infof("get file by storage id: %v failed", fatherID)
		return nil, 0, err
	}
	count, err = f.file.CountByCondition(ctx, cond)
	if err != nil {
		log.WithError(err).Infof("get count by storage id: %v failed", fatherID)
		return nil, 0, err
	}
	return byFolderID, count, nil
}

// GetFileByType 获取类型为fileType的文件
func (f *FileService) GetFileByType(ctx context.Context, storageID int64, fatherID *int64, fileType global.FileType, cursor int64, limit int) ([]*do.File, int64, error) {
	log := logger.GetLogger(ctx, "GetFileByType")
	cond := &dto.QueryCondition{StorageID: storageID, FatherID: fatherID, Type: &fileType, Cursor: cursor, Limit: limit}
	byType, err := f.file.SelectFileByCondition(ctx, cond)
	if err != nil {
		log.WithError(err).Infof("get file by type: %v failed", fileType)
		return nil, 0, err
	}
	count, err := f.file.CountByCondition(ctx, cond)
	if err != nil {
		log.WithError(err).Infof("get count by type: %v failed", fileType)
		return nil, 0, err
	}
	return byType, count, nil
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
func (f *FileService) NewFile(ctx context.Context, storageID, fatherID int64, name string) (file *do.File, err error) {
	log := logger.GetLogger(ctx, "StoreFile")

	fileInfo := &do.FileInfo{
		Hash: hasher.GetSha256().Hash(storageID, fatherID, name),
		Name: name,
	}
	if err := f.info.InsertFileInfo(ctx, nil, fileInfo); err != nil {
		log.WithError(err).Warnf("save file info: %+v failed", fileInfo)
		return nil, err
	}
	file, err = f.StoreFile(ctx, fileInfo, storageID, fatherID, name)
	if err != nil {
		log.WithError(err).Warnf("store file: %s failed", name)
		return nil, err
	}
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
func (f *FileService) StoreFile(ctx context.Context, info *do.FileInfo, storageID, fatherID int64, name string) (*do.File, error) {
	log := logger.GetLogger(ctx, "StoreFile")
	filename, err := f.FixFilename(ctx, name, storageID, fatherID)
	if err != nil {
		log.WithError(err).Warnf("fix name: %+v failed", name)
		return nil, err
	}
	file := &do.File{
		Name:      filename,
		StorageID: storageID,
		FatherID:  global.DefaultFatherID,
		Type:      int(utils.GetFileType(strings.ToLower(path.Ext(filename)))),
		FileInfo:  *info,
	}
	if len(info.OssKey) == 0 {
		file.Type = int(global.FileFolderType)
	}
	if fatherID != 0 {
		file.FatherID = fatherID
	}
	err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := f.file.InsertFile(ctx, tx, file); err != nil {
			log.WithError(err).Warnf("save file: %+v failed", file)
			return err
		}
		if info.Size != 0 {
			if err := f.storage.UpdateCurrentSize(ctx, tx, storageID, info.Size); err != nil {
				log.WithError(err).Warnf("update storage: %v, size: %v failed", storageID, info.Size)
				return err
			}
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
func (f *FileService) DeleteFile(ctx context.Context, fileID int64) error {
	log := logger.GetLogger(ctx, "DeleteFile")
	file, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Infof("get file by id: %d failed", fileID)
		return err
	}
	if err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error { return f.deleteFile(ctx, tx, file) }); err != nil {
		log.WithError(err).Warn("transaction failed")
		return err
	}
	log.Infof("delete all file by id: %v success", fileID)
	return nil
}

// deleteFile 递归删除文件信息，只涉及file表
func (f *FileService) deleteFile(ctx context.Context, tx *gorm.DB, file *do.File) error {
	log := logger.GetLogger(ctx, "deleteFile")
	if err := f.file.DeleteFileByID(ctx, tx, file.ID); err != nil {
		log.WithError(err).Warnf("delete file: %v failed", file.ID)
		return err
	}
	if file.FileInfo.Size != 0 {
		if err := f.storage.UpdateCurrentSize(ctx, tx, file.StorageID, -file.FileInfo.Size); err != nil {
			log.WithError(err).Warnf("update storage: %v size: %v failed", file.StorageID, -file.FileInfo.Size)
			return err
		}
	}
	if err := f.info.UpdateFileInfo(ctx, tx, file.FileInfoID, -1); err != nil {
		log.WithError(err).Warnf("update file info: %v failed", file.FileInfoID)
		return err
	}

	if file.Type == int(global.FileFolderType) {
		files, err := f.ScanFileByFatherID(ctx, file.StorageID, file.ID)
		if err != nil {
			log.WithError(err).Infof("get file by father id: %d failed", file.ID)
			return err
		}
		for _, child := range files {
			if err := f.deleteFile(ctx, tx, child); err != nil {
				log.WithError(err).Infof("delete file by id: %d failed", file.ID)
				return err
			}
		}
	}
	log.Infof("delete file: %v success", file.ID)
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
func (f *FileService) CopyFile(ctx context.Context, fatherID, fileID int64) error {
	log := logger.GetLogger(ctx, "CopyFile")
	file, err := f.file.SelectFileByID(ctx, fileID)
	if err != nil {
		log.WithError(err).Warnf("get file by id: %v failed", fileID)
		return err
	}
	filename, err := f.FixFilename(ctx, file.Name, file.StorageID, fatherID)
	if err != nil {
		log.WithError(err).Warnf("fix name: %+v failed", file.Name)
		return err
	}

	file.Name = filename
	if err = pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error { return f.copyFile(ctx, tx, file.StorageID, fatherID, file) }); err != nil {
		log.WithError(err).Warn("transaction failed")
		return err
	}

	log.Infof("store file: %v to father: %v of info: %+v success", fileID, fatherID, file.FileInfo)
	return nil
}

// copyFile 递归复制文件
func (f *FileService) copyFile(ctx context.Context, tx *gorm.DB, storageID, fatherID int64, file *do.File) (err error) {
	log := logger.GetLogger(ctx, "copyFile")

	var files []*do.File
	if file.Type == int(global.FileFolderType) {
		files, err = f.ScanFileByFatherID(ctx, file.StorageID, file.ID)
		if err != nil {
			log.WithError(err).Infof("get file by father id: %d failed", file.ID)
			return err
		}
	}
	file.ID = utils.GetSnowflake().Generate()
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
	file.FatherID = fatherID

	if err := f.file.InsertFile(ctx, tx, file); err != nil {
		log.WithError(err).Warnf("save file: %+v failed", file)
		return err
	}
	if file.FileInfo.Size != 0 {
		if err := f.storage.UpdateCurrentSize(ctx, tx, storageID, file.FileInfo.Size); err != nil {
			log.WithError(err).Warnf("update storage: %v, size: %v failed", storageID, file.FileInfo.Size)
			return err
		}
	}
	if err := f.info.UpdateFileInfo(ctx, tx, file.FileInfo.ID, 1); err != nil {
		log.WithError(err).Warnf("update file info: %v failed", file.FileInfo.ID)
		return err
	}
	for _, child := range files {
		if err := f.copyFile(ctx, tx, storageID, file.ID, child); err != nil {
			log.WithError(err).Infof("delete file by id: %d failed", file.ID)
			return err
		}
	}
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

// ScanFileByFatherID 获取所有FatherID下的文件
func (f *FileService) ScanFileByFatherID(ctx context.Context, storageID, fatherID int64) ([]*do.File, error) {
	var cursor, limit = int64(0), 500
	var files = make([]*do.File, 0)
	cond := &dto.QueryCondition{StorageID: storageID, FatherID: &fatherID, Limit: limit}
	for {
		cond.Cursor = cursor
		fileByFatherID, err := f.file.SelectFileByCondition(ctx, cond)
		if err != nil {
			return nil, err
		}
		for _, file := range fileByFatherID {
			cursor = file.ID
			files = append(files, file)
		}
		if len(fileByFatherID) < limit {
			return files, nil
		}
	}
}
