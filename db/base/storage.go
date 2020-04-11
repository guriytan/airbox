package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 数据仓库数据库操作接口
type StorageDao interface {
	InsertStorage(db *gorm.DB, storage *model.Storage) error

	DeleteStorageByID(db *gorm.DB, id string) error

	UpdateStorage(db *gorm.DB, storage *model.Storage) error
	UpdateCurrentSize(db *gorm.DB, id string, size int64) error
	UpdateMaxSize(db *gorm.DB, id string, size int64) error

	SelectStorageByUserID(db *gorm.DB, id string) (*model.Storage, error)
}
