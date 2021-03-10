package base

import (
	"context"

	"airbox/model"

	"gorm.io/gorm"
)

// 数据仓库数据库操作接口
type StorageDao interface {
	InsertStorage(ctx context.Context, tx *gorm.DB, storage *model.Storage) error

	DeleteStorageByID(ctx context.Context, tx *gorm.DB, id string) error

	UpdateStorage(ctx context.Context, storage *model.Storage) error
	UpdateCurrentSize(ctx context.Context, tx *gorm.DB, id string, size int64) error
	UpdateMaxSize(ctx context.Context, id string, size int64) error

	SelectStorageByUserID(ctx context.Context, id string) (*model.Storage, error)
}
