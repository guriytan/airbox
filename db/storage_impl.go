package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/model"

	"gorm.io/gorm"
)

// 数据仓库数据库操作实体
type StorageDaoImpl struct {
	db *gorm.DB
}

// InsertStorage 新增数据仓库
func (s *StorageDaoImpl) InsertStorage(ctx context.Context, tx *gorm.DB, storage *model.Storage) error {
	if tx == nil {
		tx = s.db.WithContext(ctx)
	}
	return tx.Create(storage).Error
}

// DeleteStorageByID 根据数据仓库ID删除数据仓库
func (s *StorageDaoImpl) DeleteStorageByID(ctx context.Context, tx *gorm.DB, id string) error {
	if tx == nil {
		tx = s.db.WithContext(ctx)
	}
	return tx.Delete(&model.Storage{}, "id = ?", id).Error
}

// UpdateStorage 更新数据仓库信息
func (s *StorageDaoImpl) UpdateStorage(ctx context.Context, storage *model.Storage) error {
	return s.db.WithContext(ctx).Model(&model.Storage{}).Updates(storage).Error
}

// UpdateCurrentSize 更新数据仓库最大容量
func (s *StorageDaoImpl) UpdateCurrentSize(ctx context.Context, tx *gorm.DB, id string, size int64) error {
	if tx == nil {
		tx = s.db.WithContext(ctx)
	}
	return tx.Model(&model.Storage{}).Where("id = ?", id).UpdateColumn("current_size", gorm.Expr("current_size + ?", size)).Error
}

// UpdateMaxSize 更新数据仓库当前容量
func (s *StorageDaoImpl) UpdateMaxSize(ctx context.Context, id string, size int64) error {
	return s.db.WithContext(ctx).Model(&model.Storage{}).Where("id = ?", id).UpdateColumn("max_size", gorm.Expr("max_size + ?", size)).Error
}

// SelectStorageByUserID 根据用户ID获得数据仓库
func (s *StorageDaoImpl) SelectStorageByUserID(ctx context.Context, id string) (*model.Storage, error) {
	storage := &model.Storage{}
	err := s.db.WithContext(ctx).Find(storage, "id = ?", id).Error
	return storage, err
}

var (
	storageDao     base.StorageDao
	storageDaoOnce sync.Once
)

func GetStorageDao(db *gorm.DB) base.StorageDao {
	storageDaoOnce.Do(func() {
		storageDao = &StorageDaoImpl{db: db}
	})
	return storageDao
}
