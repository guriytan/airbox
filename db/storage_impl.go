package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/model/do"

	"gorm.io/gorm"
)

// StorageDaoImpl 数据仓库数据库操作实体
type StorageDaoImpl struct {
	db *gorm.DB
}

// InsertStorage 新增数据仓库
func (s *StorageDaoImpl) InsertStorage(ctx context.Context, tx *gorm.DB, storage *do.Storage) error {
	if tx == nil {
		tx = s.db.WithContext(ctx)
	}
	return tx.Create(storage).Error
}

// DeleteStorageByID 根据数据仓库ID删除数据仓库
func (s *StorageDaoImpl) DeleteStorageByID(ctx context.Context, tx *gorm.DB, storageID int64) error {
	if tx == nil {
		tx = s.db.WithContext(ctx)
	}
	return tx.Delete(&do.Storage{}, "id = ?", storageID).Error
}

// UpdateStorage 更新数据仓库信息
func (s *StorageDaoImpl) UpdateStorage(ctx context.Context, storage *do.Storage) error {
	return s.db.WithContext(ctx).Model(&do.Storage{}).Updates(storage).Error
}

// UpdateCurrentSize 更新数据仓库最大容量
func (s *StorageDaoImpl) UpdateCurrentSize(ctx context.Context, tx *gorm.DB, storageID, size int64) error {
	if tx == nil {
		tx = s.db.WithContext(ctx)
	}
	return tx.Model(&do.Storage{}).Where("id = ?", storageID).UpdateColumn("current_size", gorm.Expr("current_size + ?", size)).Error
}

// UpdateMaxSize 更新数据仓库当前容量
func (s *StorageDaoImpl) UpdateMaxSize(ctx context.Context, storageID, size int64) error {
	return s.db.WithContext(ctx).Model(&do.Storage{}).Where("id = ?", storageID).UpdateColumn("max_size", gorm.Expr("max_size + ?", size)).Error
}

// SelectStorageByUserID 根据用户ID获得数据仓库
func (s *StorageDaoImpl) SelectStorageByUserID(ctx context.Context, storageID int64) (*do.Storage, error) {
	storage := &do.Storage{}
	result := s.db.WithContext(ctx).Find(storage, "id = ?", storageID)
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return storage, result.Error
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
