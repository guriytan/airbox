package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 文件夹数据库操作接口
type FolderDao interface {
	InsertFolder(db *gorm.DB, folder *model.Folder) error

	DeleteFolderByID(db *gorm.DB, id string) error
	DeleteFolderBySID(db *gorm.DB, sid string) error

	UpdateFolder(db *gorm.DB, folder *model.Folder) error

	SelectFolderByID(db *gorm.DB, id string) (*model.Folder, error)
	SelectFolderByName(db *gorm.DB, name, sid, fid string) (*model.Folder, error)
	SelectFolderByFatherID(db *gorm.DB, fid string) ([]model.Folder, error)
	SelectFolderByStorageID(db *gorm.DB, sid string) ([]model.Folder, error)
}
