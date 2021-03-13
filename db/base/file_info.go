package base

import (
	"context"

	"airbox/model/do"

	"gorm.io/gorm"
)

type FileInfoDao interface {
	InsertFileInfo(ctx context.Context, info *do.FileInfo) error
	DeleteFileInfo(ctx context.Context, tx *gorm.DB, id string) error
	UpdateFileInfo(ctx context.Context, tx *gorm.DB, id string, delta int64) error
	SelectFileInfoByID(ctx context.Context, id string) (*do.FileInfo, error)
	SelectFileInfoByHash(ctx context.Context, hash string) (*do.FileInfo, error)
}
