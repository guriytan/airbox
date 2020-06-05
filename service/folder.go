package service

import (
	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/model"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type FolderService struct {
	entity  base.FileEntityDao
	folder  base.FolderDao
	file    base.FileDao
	storage base.StorageDao
}

var folder *FolderService

func GetFolderService() *FolderService {
	if folder == nil {
		folder = &FolderService{
			entity:  db.GetFileEntityDao(),
			folder:  db.GetFolderDao(),
			file:    db.GetFileDao(),
			storage: db.GetStorageDao(),
		}
	}
	return folder
}

// GetFolderByID 获取该ID的文件夹
func (f *FolderService) GetFolderByID(id string) (folder *model.Folder, err error) {
	folder, err = f.folder.SelectFolderByID(global.DB, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return folder, nil
}

// GetFolderByIDWithPath 获取该ID的文件夹，并返回包括该文件夹的前置文件路径
func (f *FolderService) GetFolderByIDWithPath(id string) (folders []model.Folder, err error) {
	folder, err := f.folder.SelectFolderByID(global.DB, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	folders = append(folders, *folder)
	for folder.FatherID != nil {
		folder, err = f.folder.SelectFolderByID(global.DB, *folder.FatherID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		folders = append(folders, *folder)
	}

	return folders, nil
}

// GetFolderByStorageID 获取在仓库Sid下文件夹，一般用于初始显示
func (f *FolderService) GetFolderByStorageID(sid string) ([]model.Folder, error) {
	byStorageID, err := f.folder.SelectFolderByStorageID(global.DB, sid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byStorageID, nil
}

// GetFolderByFatherID 获取在父文件夹fid下文件夹
func (f *FolderService) GetFolderByFatherID(fid string) ([]model.Folder, error) {
	byFatherID, err := f.folder.SelectFolderByFatherID(global.DB, fid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byFatherID, nil
}

// AddFolder 添加文件夹，若文件夹在当前文件夹下已存在，则在名字后面增加数字
func (f *FolderService) AddFolder(name string, sid, fid string) (*model.Folder, error) {
	_, err := f.folder.SelectFolderByName(global.DB, name, sid, fid)
	i := 1
	for err == nil {
		name = name + "(" + strconv.FormatInt(int64(i), 10) + ")"
		_, err = f.folder.SelectFolderByName(global.DB, name, sid, fid)
		i++
	}
	if err != gorm.ErrRecordNotFound {
		return nil, errors.WithStack(err)
	}
	folder := &model.Folder{
		StorageID: sid,
		Name:      name,
	}
	if fid != "" {
		folder.FatherID = &fid
	}
	if err = f.folder.InsertFolder(global.DB, folder); err != nil {
		return nil, errors.WithStack(err)
	}
	return folder, nil
}

// DeleteFolder 删除文件夹，并迭代删除文件夹里面的所有文件和子文件夹
// 并在磁盘中删除所有文件
func (f *FolderService) DeleteFolder(id string) error {
	if err := f.deleteFolderDFS(id); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// deleteFolderDFS 迭代删除文件夹里面的所有文件和子文件夹
func (f *FolderService) deleteFolderDFS(fid string) error {
	files, _ := f.file.SelectFileByFolderID(global.DB, fid)
	for _, file := range files {
		err := deleteFile(f.entity, f.file, f.storage, file.ID)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	folders, _ := f.folder.SelectFolderByFatherID(global.DB, fid)
	for _, folder := range folders {
		if err := f.deleteFolderDFS(folder.ID); err != nil {
			return errors.WithStack(err)
		}
	}
	err := f.folder.DeleteFolderByID(global.DB, fid)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RenameFolder 重命名，需要判断当前文件夹下是否存在同样名字的文件夹
func (f *FolderService) RenameFolder(name, id string) error {
	folder, err := f.folder.SelectFolderByID(global.DB, id)
	if err != nil {
		return errors.WithStack(err)
	}
	var fid string
	if folder.FatherID != nil {
		fid = *folder.FatherID
	}
	if _, err := f.folder.SelectFolderByName(global.DB, name, folder.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	}
	return f.folder.UpdateFolder(global.DB, id, map[string]interface{}{"name": name})
}

// CopyFolder 复制文件夹，需要判断当前文件夹下是否存在同样名字的文件夹
// 需要查询文件夹下的子文件夹和文件并进行复制
func (f *FolderService) CopyFolder(fid, id string) error {
	folder, err := f.folder.SelectFolderByID(global.DB, id)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = f.folder.SelectFolderByName(global.DB, folder.Name, folder.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	}
	tx := global.DB.Begin()
	if err := f.copyFolderDFS(tx, folder, fid); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	return tx.Commit().Error
}

func (f *FolderService) copyFolderDFS(tx *gorm.DB, folder *model.Folder, fid string) error {
	id := folder.ID
	folder.Model, folder.FatherID = model.Model{}, nil
	if fid != "" {
		folder.FatherID = &fid
	}
	// 复制文件夹，获取新ID
	if err := f.folder.InsertFolder(tx, folder); err != nil {
		return errors.WithStack(err)
	}
	fileByFolderID, err := f.file.SelectFileByFolderID(tx, id) // 以folder的原ID作为父文件id的文件列表
	if err != nil {
		return errors.WithStack(err)
	}
	for _, file := range fileByFolderID {
		if _, err = insertFile(f.entity, f.file, f.storage, &file.FileEntity, file.StorageID, folder.ID); err != nil {
			return errors.WithStack(err)
		}
	}
	folderByFatherID, err := f.folder.SelectFolderByFatherID(tx, id) // 以folder的原ID作为父文件id的文件夹列表
	if err != nil {
		return errors.WithStack(err)
	}
	for _, ff := range folderByFatherID {
		// 将新文件夹的新ID作为fid，对每一个文件迭代复制
		if err := f.copyFolderDFS(tx, &ff, folder.ID); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// MoveFolder 移动文件夹，需要判断当前文件夹下是否存在同样名字的文件夹
func (f *FolderService) MoveFolder(fid, id string) error {
	folder, err := f.folder.SelectFolderByID(global.DB, id)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = f.folder.SelectFolderByName(global.DB, folder.Name, folder.StorageID, fid); err == nil {
		return errors.New(global.ErrorOfConflict)
	} else if err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	save := make(map[string]interface{})
	if fid != "" {
		save["father_id"] = fid
	} else {
		save["father_id"] = nil
	}
	return f.folder.UpdateFolder(global.DB, id, save)
}

func (f *FolderService) CreateFolder(filepath, sid, fid string) (string, error) {
	// 新建文件对应的文件夹
	if filepath != "" {
		// 分割文件夹路径
		split, query := strings.Split(filepath, "/"), true
		for _, p := range split {
			// 查询该层文件夹路径是否存在，若存在则直接进行下一层
			if query {
				if folder, err := f.folder.SelectFolderByName(global.DB, p, sid, fid); err == nil {
					fid = folder.ID
					continue
				} else if err != gorm.ErrRecordNotFound {
					return "", errors.WithStack(err)
				}
			}
			// 若不存在，则新建文件夹
			// 由于该层文件夹是新建的，因此在该层文件夹之后的文件夹都不可能会存在
			// 因此不需要再查询文件夹
			folder := &model.Folder{StorageID: sid, Name: p}
			if fid != "" {
				folder.FatherID = &fid
			}
			if err := f.folder.InsertFolder(global.DB, folder); err != nil {
				return "", errors.WithStack(err)
			}
			fid = folder.ID
			query = false
		}
	}
	return fid, nil
}
