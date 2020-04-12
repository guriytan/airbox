package service

import (
	. "airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
	"os"
	"strconv"
)

type FolderService struct {
	folder  base.FolderDao
	file    base.FileDao
	storage base.StorageDao
}

var folder *FolderService

func GetFolderService() *FolderService {
	if folder == nil {
		folder = &FolderService{
			folder:  db.GetFolderDao(),
			file:    db.GetFileDao(),
			storage: db.GetStorageDao(),
		}
	}
	return folder
}

// 删除文件夹，并迭代删除文件夹里面的所有文件和子文件夹
// 并在磁盘中删除所有文件
func (f *FolderService) DeleteFolder(id string) error {
	tx := DB.Begin()
	if err := f.deleteFolderDFS(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// 迭代删除文件夹里面的所有文件和子文件夹
func (f *FolderService) deleteFolderDFS(tx *gorm.DB, fid string) error {
	files, err := f.file.SelectFileByFolderID(tx, fid)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := f.file.DeleteFileByID(tx, file.ID); err != nil {
			return err
		}
		if err := f.storage.UpdateCurrentSize(tx, file.StorageID, int64(-file.Size)); err != nil {
			return err
		}
		_ = os.RemoveAll(file.Location)
	}
	folders, err := f.folder.SelectFolderByFatherID(tx, fid)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if err = f.deleteFolderDFS(tx, folder.ID); err != nil {
			return err
		}
	}
	return f.folder.DeleteFolderByID(tx, fid)
}

// 添加文件夹，若文件夹在当前文件夹下已存在，则在名字后面增加数字
func (f *FolderService) AddFolder(name string, sid, fid string) (*model.Folder, error) {
	_, err := f.folder.SelectFolderByName(DB, name, sid, fid)
	i := 1
	for err == nil {
		_, err = f.folder.SelectFolderByName(DB, name+"("+strconv.FormatInt(int64(i), 10)+")", sid, fid)
		i++
	}
	folder := &model.Folder{
		StorageID: sid,
		Name:      name,
	}
	if fid != "" {
		folder.FatherID = &fid
	}
	if err = f.folder.InsertFolder(DB, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// 重命名，需要判断当前文件夹下是否存在同样名字的文件夹
func (f *FolderService) Rename(name string, id, sid, fid string) error {
	if _, err := f.folder.SelectFolderByName(DB, name, sid, fid); err == nil {
		return err
	}
	return f.folder.UpdateFolder(DB, &model.Folder{
		Model: model.Model{ID: id},
		Name:  name,
	})
}

// 获取该ID的文件夹，并返回该文件夹下的子文件夹和文件
func (f *FolderService) GetFolderByID(id string) (folders []model.Folder, err error) {
	folder, err := f.folder.SelectFolderByID(DB, id)
	if err != nil {
		return
	}
	folders = append(folders, *folder)
	for folder.FatherID != nil {
		folder, err = f.folder.SelectFolderByID(DB, *folder.FatherID)
		if err != nil {
			return nil, err
		}
		folders = append(folders, *folder)
	}

	return folders, nil
}

// 获取在仓库Sid下文件夹，一般用于初始显示
func (f *FolderService) GetFolderByStorageID(sid string) ([]model.Folder, error) {
	return f.folder.SelectFolderByStorageID(DB, sid)
}

func (f *FolderService) GetFolderByFatherID(fid string) ([]model.Folder, error) {
	return f.folder.SelectFolderByFatherID(DB, fid)
}
