package service

import (
	. "airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/model"
	"airbox/utils"
	"errors"
	"github.com/google/uuid"
	"mime/multipart"
	"os"
	"path"
	"strings"
)

type FileService struct {
	file    base.FileDao
	storage base.StorageDao
}

// GetFileByID 获取文件信息，用于下载
func (f *FileService) GetFileByID(id string) (*model.File, error) {
	return f.file.SelectFileByID(DB, id)
}

// GetFileByStorageID 获取在仓库Sid下的文件，一般用于初始显示
func (f *FileService) GetFileByStorageID(sid string) ([]model.File, error) {
	return f.file.SelectFileByStorageID(DB, sid)
}

// SelectFileByFolderID 获取在父文件夹fid下的文件
func (f *FileService) SelectFileByFolderID(fid string) (files []model.File, err error) {
	return f.file.SelectFileByFolderID(DB, fid)
}

// GetFileByType 获取类型为t的文件
func (f *FileService) GetFileByType(t string) ([]model.File, error) {
	return f.file.SelectFileByType(DB, t)
}

// SelectFileTypeCount 获取不同类型文件的统计数量
func (f *FileService) SelectFileTypeCount(sid string) (types []model.Statistics, err error) {
	return f.file.SelectFileTypeCount(DB, sid)
}

// UploadFile 保存文件信息，并更新数据仓库的容量大小
func (f *FileService) UploadFile(filepath, sid, fid string, part *multipart.Part, size uint64) (string, error) {
	filename := part.FileName()
	// 由于在上传文件夹模型下Filename()将有前置文件夹路径。因此统一剪切
	filename = filename[strings.LastIndex(filename, "/")+1:]
	if filename == "" {
		return "", errors.New("without file")
	}
	// 判断是否已存在同名文件并修改文件名（增加数字编号）
	i := 1
	for {
		if _, err := f.file.SelectFileByName(DB, filename, sid, fid); err != nil {
			break
		}
		filename = utils.AddIndexToFilename(filename, i)
		i++
	}
	// 调用Upload上传并返回文件长度
	if err := utils.Upload(filepath, filename, part, size); err != nil {
		return "", err
	}
	return filename, nil
}

// StoreFile 保存文件信息并更新仓库现容量大小
func (f *FileService) StoreFile(md5, filename, filepath string, size uint64, sid, fid string) (*model.File, error) {
	suffix := strings.ToLower(path.Ext(filename))
	file := &model.File{
		Name:      filename,
		Size:      size,
		Hash:      md5,
		Location:  filepath,
		StorageID: sid,
		Suffix:    suffix,
		Type:      int(utils.GetFileType(suffix)),
	}
	if fid != "" {
		file.FolderID = &fid
	}
	// 匿名函数主要用于判断错误并执行删除文件操作
	store := func() error {
		tx := DB.Begin()
		if err := f.file.InsertFile(tx, file); err != nil {
			tx.Rollback()
			return err
		}
		if err := f.storage.UpdateCurrentSize(tx, sid, int64(size)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit().Error; err != nil {
			return err
		}
		return nil
	}
	if err := store(); err != nil {
		_ = os.RemoveAll(filepath)
		return nil, err
	}
	return file, nil
}

// DeleteFile 删除文件信息，并更新数据仓库的容量大小
func (f *FileService) DeleteFile(id string) error {
	file, err := f.file.SelectFileByID(DB, id)
	if err != nil {
		return err
	}
	tx := DB.Begin()
	if err := f.file.DeleteFileByID(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := f.storage.UpdateCurrentSize(tx, file.StorageID, int64(-file.Size)); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return os.RemoveAll(file.Location)
}

// RenameFile 重命名，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) RenameFile(name, id string) error {
	file, err := f.file.SelectFileByID(DB, id)
	if err != nil {
		return err
	}
	var fid string // go 不支持三元表达式，因此需要额外处理
	if file.FolderID != nil {
		fid = *file.FolderID
	}
	if _, err = f.file.SelectFileByName(DB, name, file.StorageID, fid); err == nil {
		return errors.New(ErrorOfConflict)
	}
	if err := f.file.UpdateFile(DB, id, map[string]interface{}{"name": name}); err != nil {
		return err
	}
	return os.Rename(file.Location+file.Name, file.Location+name)
}

// MoveFile 移动文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) MoveFile(fid, id string) error {
	file, err := f.file.SelectFileByID(DB, id)
	if err != nil {
		return err
	}
	if _, err = f.file.SelectFileByName(DB, file.Name, file.StorageID, fid); err == nil {
		return errors.New(ErrorOfConflict)
	}
	save := make(map[string]interface{})
	if fid != "" {
		save["folder_id"] = fid
	} else {
		save["folder_id"] = nil
	}
	return f.file.UpdateFile(DB, id, save)
}

// CopyFile 复制文件，需要判断当前文件夹下是否存在同样名字的文件
func (f *FileService) CopyFile(fid, id string) error {
	file, err := f.file.SelectFileByID(DB, id)
	if err != nil {
		return err
	}
	if _, err = f.file.SelectFileByName(DB, file.Name, file.StorageID, fid); err == nil {
		return errors.New(ErrorOfConflict)
	}
	filepath := Env.Upload.Dir + "/" + file.StorageID + "/" + uuid.New().String() + "/"
	_ = os.MkdirAll(filepath, os.ModePerm)
	if _, err := f.StoreFile(file.Hash, file.Name, filepath, file.Size, file.StorageID, fid); err != nil {
		return err
	}
	if _, err := utils.CopyFile(filepath+file.Name, file.Location+file.Name); err != nil {
		return err
	}
	return nil
}

var file *FileService

func GetFileService() *FileService {
	if file == nil {
		file = &FileService{
			file:    db.GetFileDao(),
			storage: db.GetStorageDao(),
		}
	}
	return file
}
