package service

import (
	. "airbox/config"
	"airbox/db"
	"airbox/db/base"
	f "airbox/file"
	"airbox/model"
	"airbox/utils"
	"fmt"
	"github.com/google/uuid"
	"mime/multipart"
	"os"
	"path"
	"strings"
)

type FileService struct {
	file    base.FileDao
	storage base.StorageDao
	folder  *FolderService
	user    base.UserDao
}

// 保存文件信息，并更新数据仓库的容量大小
func (s *FileService) UploadFile(part *multipart.Part, filename string, size uint64, id, fid string) error {
	user, err := s.user.SelectUserByID(id)
	if err != nil {
		return err
	} else if user.Storage.CurrentSize+size >= user.Storage.MaxSize {
		return ErrorOutOfSpace
	}
	// 计算文件实际存储路径
	filepath := PrefixMasterDirectory + user.Storage.ID + "/" + uuid.New().String() + "/"
	_ = os.MkdirAll(filepath, os.ModePerm)
	// 判断是否已存在同名文件并修改文件名（增加数字编号）
	i := 1
	for {
		if _, err := s.file.SelectFileByName(DB, filename, user.Storage.ID, &fid); err != nil {
			break
		}
		filename = utils.AddIndexToFilename(filename, i)
		i++
	}
	// 调用Upload上传并返回文件长度
	if err := f.Upload(filepath, filename, part, size); err != nil {
		return err
	}
	// 计算文件hash
	hash, err := utils.SHA256Sum(filepath + filename)
	if err != nil {
		fmt.Println(err)
	}
	if err := s.storeFile(hash, filename, filepath, size, user.Storage.ID, fid); err != nil {
		if err = os.RemoveAll(filepath); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileService) storeFile(hash, filename, filepath string, size uint64, sid, fid string) error {
	suffix := strings.ToLower(path.Ext(filename))
	file := &model.File{
		Hash:      hash,
		Name:      filename,
		Size:      size,
		Location:  filepath,
		StorageID: sid,
		Suffix:    suffix,
		Type:      int(utils.GetFileType(suffix)),
	}
	if fid != "" {
		file.FolderID = &fid
	}
	tx := DB.Begin()
	if err := s.file.InsertFile(tx, file); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.storage.UpdateCurrentSize(tx, sid, int64(size)); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *FileService) GetFileByID(id string) (*model.File, error) {
	return s.file.SelectFileByID(DB, id)
}

// 删除文件信息，并更新数据仓库的容量大小
func (s *FileService) DeleteFile(id string) error {
	file, err := s.file.SelectFileByID(DB, id)
	if err != nil {
		return err
	}
	tx := DB.Begin()
	if err := s.file.DeleteFileByID(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.storage.UpdateCurrentSize(tx, file.StorageID, int64(-file.Size)); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return os.RemoveAll(file.Location)
}

// 重命名，需要判断当前文件夹下是否存在同样名字的文件
func (s *FileService) Rename(name, id string) error {
	file, err := s.file.SelectFileByID(DB, id)
	if err != nil {
		return err
	}
	_, err = s.file.SelectFileByName(DB, name, file.StorageID, file.FolderID)
	if err == nil {
		return err
	}
	if err := s.file.UpdateFile(DB, &model.File{
		Model: model.Model{ID: id},
		Name:  name,
	}); err != nil {
		return err
	}
	return os.Rename(file.Location+file.Name, file.Location+name)
}

// 获取在仓库Sid下的文件，一般用于初始显示
func (s *FileService) GetFileByStorageID(sid string) ([]model.File, error) {
	return s.file.SelectFileByStorageID(DB, sid)
}

// 获取类型为t的文件
func (s *FileService) GetFileByType(t string) ([]model.File, error) {
	return s.file.SelectFileByType(DB, t)
}

// 获取在父文件夹fid下的文件
func (s *FileService) SelectFileByFolderID(fid string) (files []model.File, err error) {
	return s.file.SelectFileByFolderID(DB, fid)
}

func (s *FileService) SelectFileTypeCount() (types []model.Statistics, err error) {
	return s.file.SelectFileTypeCount(DB)
}

var file *FileService

func GetFileService() *FileService {
	if file == nil {
		file = &FileService{
			file:    db.GetFileDao(),
			storage: db.GetStorageDao(),
			folder:  GetFolderService(),
		}
	}
	return file
}
