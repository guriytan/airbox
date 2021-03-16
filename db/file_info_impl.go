package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/model/do"

	"gorm.io/gorm"
)

// 文件数据库操作实体
type FileInfoDaoImpl struct {
	db *gorm.DB
}

// InsertFileInfo 新增文件信息
func (f *FileInfoDaoImpl) InsertFileInfo(ctx context.Context, info *do.FileInfo) error {
	return f.db.WithContext(ctx).Create(info).Error
}

// DeleteFileInfo 根据文件信息ID删除文件信息
func (f *FileInfoDaoImpl) DeleteFileInfo(ctx context.Context, tx *gorm.DB, infoID string) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Delete(&do.FileInfo{}, "id = ?", infoID).Error
}

// UpdateFileInfo 更新文件信息
func (f *FileInfoDaoImpl) UpdateFileInfo(ctx context.Context, tx *gorm.DB, infoID string, delta int64) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Model(&do.FileInfo{}).Where("id = ?", infoID).UpdateColumn("link", gorm.Expr("link + ?", delta)).Error
}

// SelectFileInfoByID 根据文件ID获得文件信息
func (f *FileInfoDaoImpl) SelectFileInfoByID(ctx context.Context, infoID string) (*do.FileInfo, error) {
	info := &do.FileInfo{}
	err := f.db.WithContext(ctx).Find(info, "id = ?", infoID).Error
	return info, err
}

// SelectFileInfoByHash 根据Hash获得文件信息
func (f *FileInfoDaoImpl) SelectFileInfoByHash(ctx context.Context, hash string) (*do.FileInfo, error) {
	info := &do.FileInfo{}
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
