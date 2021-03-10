package base

import (
	"context"

	"airbox/model"

	"gorm.io/gorm"
)

// 文件数据库操作接口
type FileDao interface {
	InsertFile(ctx context.Context, tx *gorm.DB, file *model.File) error

	DeleteFileByID(ctx context.Context, id string) error
	DeleteFileByStorageID(ctx context.Context, tx *gorm.DB, storageID string) error

	UpdateFile(ctx context.Context, id string, file map[string]interface{}) error

	SelectFileByID(ctx context.Context, id string) (*model.File, error)
	SelectFileByName(ctx context.Context, name, storageID, fatherID string) (*model.File, error)
	SelectFileByFatherID(ctx context.Context, fatherID string) ([]*model.File, error)
	SelectFileByStorageID(ctx context.Context, storageID string) ([]*model.File, error)
	SelectFileByType(ctx context.Context, fileType string) ([]*model.File, error)
	SelectFileTypeCount(ctx context.Context, storageID string) (types []*model.Statistics, err error)
}
