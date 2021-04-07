package base

import (
	"context"

	"airbox/model/do"

	"gorm.io/gorm"
)

// 文件数据库操作接口
type FileDao interface {
	InsertFile(ctx context.Context, tx *gorm.DB, file *do.File) error

	DeleteFileByID(ctx context.Context, tx *gorm.DB, fileID int64) error
	DeleteFileByStorageID(ctx context.Context, tx *gorm.DB, storageID int64) error

	UpdateFile(ctx context.Context, fileID int64, file map[string]interface{}) error

	SelectFileByID(ctx context.Context, fileID int64) (*do.File, error)
	SelectFileByName(ctx context.Context, name string, storageID, fatherID int64) ([]*do.File, error)
	SelectFileByFatherIDAndType(ctx context.Context, fatherID int64, fileType []int, cursor int64, limit int) ([]*do.File, error)
	SelectFileTypeCount(ctx context.Context, storageID int64) (types []*do.Statistics, err error)
}
