package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/model"

	"gorm.io/gorm"
)

// 文件数据库操作实体
type FileInfoDaoImpl struct {
	db *gorm.DB
}

func (f *FileInfoDaoImpl) InsertFileInfo(ctx context.Context, info *model.FileInfo) error {
	return f.db.WithContext(ctx).Create(info).Error
}

func (f *FileInfoDaoImpl) DeleteFileInfo(ctx context.Context, tx *gorm.DB, id string) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Delete(&model.FileInfo{}, "id = ?", id).Error
}

func (f *FileInfoDaoImpl) UpdateFileInfo(ctx context.Context, tx *gorm.DB, id string, delta int64) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Model(&model.FileInfo{}).Where("file_info_id = ?", id).UpdateColumn("link", gorm.Expr("link + ?", delta)).Error
}

func (f *FileInfoDaoImpl) SelectFileInfoByID(ctx context.Context, id string) (*model.FileInfo, error) {
	info := &model.FileInfo{}
	err := f.db.WithContext(ctx).Find(info, "id = ?", id).Error
	return info, err
}

func (f *FileInfoDaoImpl) SelectFileInfoByHash(ctx context.Context, hash string) (*model.FileInfo, error) {
	info := &model.FileInfo{}
	err := f.db.WithContext(ctx).Find(info, "hash = ?", hash).Error
	return info, err
}

var (
	infoDao     base.FileInfoDao
	infoDaoOnce sync.Once
)

func GetFileInfoDao(db *gorm.DB) base.FileInfoDao {
	infoDaoOnce.Do(func() {
		infoDao = &FileInfoDaoImpl{db: db}
	})
	return infoDao
}
