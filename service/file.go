package service

import (
	"airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/model"
	"airbox/utils"
	"airbox/utils/disk"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"mime/multipart"
	"os"
	"path"
	"strings"
)

type FileService struct {
	file    base.FileDao
	info    base.FileInfoDao
	storage base.StorageDao
}

// GetFileByID 获取文件信息，用于下载
func (f *FileService) GetFileByID(id string) (*model.File, error) {
	fileByID, err := f.file.SelectFileByID(global.DB, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return fileByID, nil
}

// GetFileByStorageID 获取在仓库Sid下的文件，一般用于初始显示
func (f *FileService) GetFileByStorageID(sid string) ([]model.File, error) {
	byStorageID, err := f.file.SelectFileByStorageID(global.DB, sid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byStorageID, nil
}

// SelectFileByFolderID 获取在父文件夹fid下的文件
func (f *FileService) SelectFileByFolderID(fid string) (files []model.File, err error) {
	byFolderID, err := f.file.SelectFileByFolderID(global.DB, fid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byFolderID, nil
}

// GetFileByType 获取类型为t的文件
func (f *FileService) GetFileByType(t string) ([]model.File, error) {
	byType, err := f.file.SelectFileByType(global.DB, t)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byType, nil
}

// SelectFileTypeCount 获取不同类型文件的统计数量
func (f *FileService) SelectFileTypeCount(sid string) (types []model.Statistics, err error) {
	typeCount, err := f.file.SelectFileTypeCount(global.DB, sid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return typeCount, nil
}

// UploadFile 保存文件信息，并更新数据仓库的容量大小
func (f *FileService) UploadFile(part *multipart.Part, sid, md5 string, size uint64) (*model.FileInfo, error) {
	// 计算文件实际存储路径
	filepath := config.Env.Upload.Dir + "/" + sid + "/" + uuid.New().String() + "/"
	filename := part.FileName()
	// 由于在上传文件夹模型下Filename()将有前置文件夹路径。因此统一剪切
	filename = filename[strings.LastIndex(filename, "/")+1:]
	if filename == "" {
		return nil, errors.New("without file")
	}
	// 调用Upload上传并返回文件长度
	if err := disk.Upload(filepath, filename, part, size); err != nil {
		return nil, errors.WithStack(err)
	}
	info := &model.FileInfo{
		Hash: md5,
		Name: filename,
		Size: size,
		Path: filepath,
	}
	tx := global.DB.Begin()
	err := f.info.InsertFileInfo(tx, info)
	if err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	err = f.info.InsertFileCount(tx, &model.FileCount{FileInfoID: info.ID})
	if err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	tx.Commit()
	return info, nil
}

// StoreFile 保存文件信息并更新仓库现容量大小
func (f *FileService) StoreFile(info *model.FileInfo, sid, fid string) (*model.File, error) {
	return insertFile(f.info, f.file, f.storage, info, sid, fid)
}

// DeleteFile 删除文件信息，并更新数据仓库的容量大小
func (f *FileService) DeleteFile(id string) error {
	return deleteFile(f.info, f.file, f.storage, id)
}

// RenameFile 重命名，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) RenameFile(name, id string) error {
	file, err := f.file.SelectFileByID(global.DB, id)
	if err != nil {
		return errors.WithStack(err)
	}
	var fid string // go 不支持三元表达式，因此需要额外处理
	if file.FolderID != nil {
		fid = *file.FolderID
	}
	if _, err = f.file.SelectFileByName(global.DB, name, file.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	} else if err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	if err := f.file.UpdateFile(global.DB, id, map[string]interface{}{"name": name}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// MoveFile 移动文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) MoveFile(fid, id string) error {
	file, err := f.file.SelectFileByID(global.DB, id)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = f.file.SelectFileByName(global.DB, file.Name, file.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	}
	save := make(map[string]interface{})
	if fid != "" {
		save["folder_id"] = fid
	} else {
		save["folder_id"] = nil
	}
	return f.file.UpdateFile(global.DB, id, save)
}

// CopyFile 复制文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) CopyFile(fid, id string) error {
	file, err := f.file.SelectFileByID(global.DB, id)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = f.StoreFile(&file.FileInfo, file.StorageID, fid)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (f *FileService) SelectFileByMD5(md5 string) (*model.FileInfo, error) {
	byMD5, err := f.info.SelectFileInfoByMD5(global.DB, md5)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byMD5, nil
}

var file *FileService

func GetFileService() *FileService {
	if file == nil {
		file = &FileService{
			info:    db.GetFileInfoDao(),
			file:    db.GetFileDao(),
			storage: db.GetStorageDao(),
		}
	}
	return file
}

func deleteFile(infoDao base.FileInfoDao, fileDao base.FileDao, storageDao base.StorageDao, id string) error {
	tx := global.DB.Begin()
	file, err := fileDao.SelectFileByID(tx, id)
	if err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := fileDao.DeleteFileByID(tx, id); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := storageDao.UpdateCurrentSize(tx, file.StorageID, int64(-file.FileInfo.Size)); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	count, err := infoDao.SelectFileCountByInfoID(tx, file.FileInfoID)
	if err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if count.Link > 1 {
		if err := infoDao.UpdateFileInfo(tx, file.FileInfoID, -1); err != nil {
			tx.Rollback()
			return errors.WithStack(err)
		}
		if err := tx.Commit().Error; err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := infoDao.DeleteFileInfo(tx, file.FileInfoID); err != nil {
			tx.Rollback()
			return errors.WithStack(err)
		}
		if err := infoDao.DeleteFileCount(tx, file.FileInfoID); err != nil {
			tx.Rollback()
			return errors.WithStack(err)
		}
		if err := tx.Commit().Error; err != nil {
			return errors.WithStack(err)
		}
		go func() {
			err = os.RemoveAll(file.FileInfo.Path)
			if err != nil {
				global.LOGGER.Printf("%s\n", err.Error())
			}
		}()
	}
	return nil
}

func insertFile(infoDao base.FileInfoDao, fileDao base.FileDao, storageDao base.StorageDao,
	info *model.FileInfo, sid, fid string) (*model.File, error) {
	filename := info.Name
	// 判断是否已存在同名文件并修改文件名（增加数字编号）
	i := 1
	for {
		if _, err := fileDao.SelectFileByName(global.DB, filename, sid, fid); err == gorm.ErrRecordNotFound {
			break
		} else if err != nil {
			return nil, errors.WithStack(err)
		}
		filename = utils.AddIndexToFilename(filename, i)
		i++
	}
	file := &model.File{
		Name:      filename,
		StorageID: sid,
		Type:      int(utils.GetFileType(strings.ToLower(path.Ext(filename)))),
		FileInfo:  *info,
	}
	if fid != "" {
		file.FolderID = &fid
	}
	tx := global.DB.Begin()
	if err := fileDao.InsertFile(tx, file); err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	if err := storageDao.UpdateCurrentSize(tx, sid, int64(info.Size)); err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	if err := infoDao.UpdateFileInfo(tx, info.ID, 1); err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return file, nil
}
