package base

import (
	"context"

	"airbox/model/do"

	"gorm.io/gorm"
)

// 文件数据库操作接口
type FileDao interface {
	InsertFile(ctx context.Context, tx *gorm.DB, file *do.File) error

	DeleteFileByID(ctx context.Context, fileID string) error
	DeleteFileByStorageID(ctx context.Context, tx *gorm.DB, storageID string) error

	UpdateFile(ctx context.Context, fileID string, file map[string]interface{}) error

	SelectFileByID(ctx context.Context, fileID string) (*do.File, error)
	SelectFileByName(ctx context.Context, name, storageID, fatherID string) (*do.File, error)
	SelectFileByFatherID(ctx context.Context, fatherID string) ([]*do.File, error)
	SelectFileByType(ctx context.Context, fileType int) ([]*do.File, error)
	SelectFileTypeCount(ctx context.Context, storageID string) (types []*do.Statistics, err error)
}
