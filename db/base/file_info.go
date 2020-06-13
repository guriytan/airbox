package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

type FileInfoDao interface {
	InsertFileInfo(db *gorm.DB, info *model.FileInfo) error
	InsertFileCount(db *gorm.DB, count *model.FileCount) error
	UpdateFileInfo(db *gorm.DB, id string, size int64) error
	DeleteFileInfo(db *gorm.DB, id string) error
	DeleteFileCount(db *gorm.DB, id string) error
	SelectFileInfoByID(db *gorm.DB, id string) (*model.FileInfo, error)
	SelectFileInfoByMD5(db *gorm.DB, md5 string) (*model.FileInfo, error)
	SelectFileCountByInfoID(db *gorm.DB, id string) (*model.FileCount, error)
}
