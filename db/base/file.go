package base

import (
	"context"

	"airbox/model/do"
	"airbox/model/dto"

	"gorm.io/gorm"
)

// FileDao 文件数据库操作接口
type FileDao interface {
	InsertFile(ctx context.Context, tx *gorm.DB, file *do.File) error

	DeleteFileByID(ctx context.Context, tx *gorm.DB, fileID int64) error
	DeleteFileByStorageID(ctx context.Context, tx *gorm.DB, storageID int64) error

	UpdateFile(ctx context.Context, fileID int64, file map[string]interface{}) error

	SelectFileByID(ctx context.Context, fileID int64) (*do.File, error)
	SelectFileByName(ctx context.Context, name string, storageID, fatherID int64) ([]*do.File, error)
	SelectFileByCondition(ctx context.Context, cond *dto.QueryCondition) ([]*do.File, error)
	CountByCondition(ctx context.Context, cond *dto.QueryCondition) (count int64, err error)
	SelectFileTypeCount(ctx context.Context, storageID int64) (types []*do.Statistics, err error)
}
