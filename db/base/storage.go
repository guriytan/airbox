package base

import (
	"context"

	"airbox/model/do"

	"gorm.io/gorm"
)

// 数据仓库数据库操作接口
type StorageDao interface {
	InsertStorage(ctx context.Context, tx *gorm.DB, storage *do.Storage) error

	DeleteStorageByID(ctx context.Context, tx *gorm.DB, storageID string) error

	UpdateStorage(ctx context.Context, storage *do.Storage) error
	UpdateCurrentSize(ctx context.Context, tx *gorm.DB, storageID string, size int64) error
	UpdateMaxSize(ctx context.Context, storageID string, size int64) error

	SelectStorageByUserID(ctx context.Context, storageID string) (*do.Storage, error)
}
