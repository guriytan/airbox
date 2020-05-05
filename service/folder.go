package service

import (
	. "airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/model"
	"airbox/utils"
	"errors"
	"github.com/google/uuid"
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

// GetFolderByID 获取该ID的文件夹，并返回该文件夹下的子文件夹和文件
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

// GetFolderByStorageID 获取在仓库Sid下文件夹，一般用于初始显示
func (f *FolderService) GetFolderByStorageID(sid string) ([]model.Folder, error) {
	return f.folder.SelectFolderByStorageID(DB, sid)
}

// GetFolderByFatherID 获取在父文件夹fid下文件夹
func (f *FolderService) GetFolderByFatherID(fid string) ([]model.Folder, error) {
	return f.folder.SelectFolderByFatherID(DB, fid)
}

// AddFolder 添加文件夹，若文件夹在当前文件夹下已存在，则在名字后面增加数字
func (f *FolderService) AddFolder(name string, sid, fid string) (*model.Folder, error) {
	_, err := f.folder.SelectFolderByName(DB, name, sid, fid)
	i := 1
	for err == nil {
		name = name + "(" + strconv.FormatInt(int64(i), 10) + ")"
		_, err = f.folder.SelectFolderByName(DB, name, sid, fid)
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

// DeleteFolder 删除文件夹，并迭代删除文件夹里面的所有文件和子文件夹
// 并在磁盘中删除所有文件
func (f *FolderService) DeleteFolder(id string) error {
	tx := DB.Begin()
	if err := f.deleteFolderDFS(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// deleteFolderDFS 迭代删除文件夹里面的所有文件和子文件夹
func (f *FolderService) deleteFolderDFS(tx *gorm.DB, fid string) error {
	files, _ := f.file.SelectFileByFolderID(tx, fid)
	for _, file := range files {
		if err := f.file.DeleteFileByID(tx, file.ID); err != nil {
			return err
		}
		if err := f.storage.UpdateCurrentSize(tx, file.StorageID, int64(-file.Size)); err != nil {
			return err
		}
		_ = os.RemoveAll(file.Location)
	}
	folders, _ := f.folder.SelectFolderByFatherID(tx, fid)
	for _, folder := range folders {
		if err := f.deleteFolderDFS(tx, folder.ID); err != nil {
			return err
		}
	}
	return f.folder.DeleteFolderByID(tx, fid)
}

// RenameFolder 重命名，需要判断当前文件夹下是否存在同样名字的文件夹
func (f *FolderService) RenameFolder(name, id string) error {
	folder, err := f.folder.SelectFolderByID(DB, id)
	if err != nil {
		return err
	}
	var fid string
	if folder.FatherID != nil {
		fid = *folder.FatherID
	}
	if _, err := f.folder.SelectFolderByName(DB, name, folder.StorageID, fid); err == nil {
		return errors.New(ErrorOfConflict)
	}
	return f.folder.UpdateFolder(DB, id, map[string]interface{}{"name": name})
}

// CopyFolder 复制文件夹，需要判断当前文件夹下是否存在同样名字的文件夹
// 需要查询文件夹下的子文件夹和文件并进行复制
func (f *FolderService) CopyFolder(fid, id string) error {
	folder, err := f.folder.SelectFolderByID(DB, id)
	if err != nil {
		return err
	}
	if _, err = f.folder.SelectFolderByName(DB, folder.Name, folder.StorageID, fid); err == nil {
		return errors.New(ErrorOfConflict)
	}
	tx := DB.Begin()
	if err := f.copyFolderDFS(tx, folder, fid); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (f *FolderService) copyFolderDFS(tx *gorm.DB, folder *model.Folder, fid string) error {
	id := folder.ID
	folder.Model = model.Model{}
	folder.FatherID = nil
	if fid != "" {
		folder.FatherID = &fid
	}
	// 复制文件夹，获取新ID
	if err := f.folder.InsertFolder(tx, folder); err != nil {
		return err
	}
	fileByFolderID, _ := f.file.SelectFileByFolderID(tx, id) // 以folder的原ID作为父文件id的文件列表
	for _, file := range fileByFolderID {
		previous := file.Location // 原有文件路径
		file.Model = model.Model{}
		file.Location = FilePrefixMasterDirectory + file.StorageID + "/" + uuid.New().String() + "/"
		_ = os.MkdirAll(file.Location, os.ModePerm)
		file.FolderID = &folder.ID // 复制文件的Fid为新文件夹的新ID
		if err := f.file.InsertFile(tx, &file); err != nil {
			return err
		}
		if err := f.storage.UpdateCurrentSize(tx, file.StorageID, int64(file.Size)); err != nil {
			return err
		}
		_, _ = utils.CopyFile(file.Location+file.Name, previous+file.Name)
	}
	folderByFatherID, _ := f.folder.SelectFolderByFatherID(tx, id) // 以folder的原ID作为父文件id的文件夹列表
	for _, ff := range folderByFatherID {
		// 将新文件夹的新ID作为fid，对每一个文件迭代复制
		if err := f.copyFolderDFS(tx, &ff, folder.ID); err != nil {
			return err
		}
	}
	return nil
}

// MoveFolder 移动文件夹，需要判断当前文件夹下是否存在同样名字的文件夹
func (f *FolderService) MoveFolder(fid, id string) error {
	folder, err := f.folder.SelectFolderByID(DB, id)
	if err != nil {
		return err
	}
	if _, err = f.folder.SelectFolderByName(DB, folder.Name, folder.StorageID, fid); err == nil {
		return errors.New(ErrorOfConflict)
	}
	save := make(map[string]interface{})
	if fid != "" {
		save["father_id"] = fid
	} else {
		save["father_id"] = nil
	}
	return f.folder.UpdateFolder(DB, id, save)
}
