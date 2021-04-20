package base

import (
	"context"

	"airbox/model/do"

	"gorm.io/gorm"
)

// FileInfoDao 文件数据库操作接口
type FileInfoDao interface {
	InsertFileInfo(ctx context.Context, tx *gorm.DB, info *do.FileInfo) error
	DeleteFileInfo(ctx context.Context, tx *gorm.DB, infoID int64) error
	UpdateFileInfo(ctx context.Context, tx *gorm.DB, infoID int64, delta int64) error
	SelectFileInfoByID(ctx context.Context, infoID int64) (*do.FileInfo, error)
	SelectFileInfoByHash(ctx context.Context, hash string) (*do.FileInfo, error)
}
