package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 文件数据库操作接口
type FileDao interface {
	InsertFile(db *gorm.DB, file *model.File) error

	DeleteFileByID(db *gorm.DB, id string) error
	DeleteFileBySID(db *gorm.DB, sid string) error

	UpdateFile(db *gorm.DB, id string, file map[string]interface{}) error

	SelectFileByID(db *gorm.DB, id string) (*model.File, error)
	SelectFileByName(db *gorm.DB, name, sid, fid string) (*model.File, error)
	SelectFileByFolderID(db *gorm.DB, fid string) ([]model.File, error)
	SelectFileByStorageID(db *gorm.DB, sid string) ([]model.File, error)
	SelectFileByType(db *gorm.DB, t string) ([]model.File, error)
	SelectFileTypeCount(db *gorm.DB, sid string) (types []model.Statistics, err error)
}
