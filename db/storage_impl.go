package db

import (
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 数据仓库数据库操作实体
type StorageDaoImpl struct {
}

// UpdateCurrentSize 更新数据仓库最大容量
func (s *StorageDaoImpl) UpdateCurrentSize(db *gorm.DB, id string, size int64) error {
	return db.Table("storage").Where("id = ?", id).UpdateColumn("current_size", gorm.Expr("current_size + ?", size)).Error
}

// UpdateMaxSize 更新数据仓库当前容量
func (s *StorageDaoImpl) UpdateMaxSize(db *gorm.DB, id string, size int64) error {
	return db.Table("storage").Where("id = ?", id).UpdateColumn("max_size", gorm.Expr("max_size + ?", size)).Error
}

// InsertStorage 新增数据仓库
func (s *StorageDaoImpl) InsertStorage(db *gorm.DB, storage *model.Storage) error {
	return db.Create(storage).Error
}

// DeleteStorageByID 根据数据仓库ID删除数据仓库
func (s *StorageDaoImpl) DeleteStorageByID(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Storage{}).Error
}

// UpdateStorage 更新数据仓库信息
func (s *StorageDaoImpl) UpdateStorage(db *gorm.DB, storage *model.Storage) error {
	return db.Table("storage").Update(storage).Error
}

// SelectStorageByUserID 根据用户ID获得数据仓库
func (s *StorageDaoImpl) SelectStorageByUserID(db *gorm.DB, id string) (*model.Storage, error) {
	storage := &model.Storage{}
	err := db.Where("id = ?", id).First(storage).Error
	return storage, err
}

var storage base.StorageDao

func GetStorageDao() base.StorageDao {
	if storage == nil {
		storage = &StorageDaoImpl{}
	}
	return storage
}
