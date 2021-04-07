package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/global"
	"airbox/model/do"

	"gorm.io/gorm"
)

// 文件数据库操作实体
type FileDaoImpl struct {
	db *gorm.DB
}

// InsertFile 新增文件
func (f *FileDaoImpl) InsertFile(ctx context.Context, tx *gorm.DB, file *do.File) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Create(file).Error
}

// DeleteFileByID 根据文件ID删除文件
func (f *FileDaoImpl) DeleteFileByID(ctx context.Context, tx *gorm.DB, fileID int64) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Delete(&do.File{}, "id = ?", fileID).Error
}

// DeleteFileByStorageID 根据仓库ID删除文件
func (f *FileDaoImpl) DeleteFileByStorageID(ctx context.Context, tx *gorm.DB, storageID int64) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Delete(&do.File{}, "storage_id = ?", storageID).Error
}

// UpdateFile 更新文件信息
func (f *FileDaoImpl) UpdateFile(ctx context.Context, fileID int64, file map[string]interface{}) error {
	return f.db.WithContext(ctx).Model(&do.File{}).Where("id = ?", fileID).Updates(file).Error
}

// SelectFileByID 根据文件ID获得文件
func (f *FileDaoImpl) SelectFileByID(ctx context.Context, fileID int64) (*do.File, error) {
	file := &do.File{}
	res := f.db.WithContext(ctx).Preload("FileInfo").Where("id = ?", fileID).Find(file)
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return file, res.Error
}

// SelectFileByType 根据文件夹ID和文件类型获得文件
func (f *FileDaoImpl) SelectFileByFatherIDAndType(ctx context.Context, fatherID int64, fileType []int, cursor int64, limit int) (files []*do.File, err error) {
	tx := f.db.WithContext(ctx).Preload("FileInfo")
	if fatherID != 0 {
		tx = tx.Where("father_id = ?", fatherID)
	}
	if len(fileType) != 0 {
		tx = tx.Where("type in ?", fileType)
	}
	err = tx.Order("updated_at desc").Find(&files).Error
	return
}

// SelectFileByName 在数据仓库下查看文件夹下是否有文件名为name的文件
func (f *FileDaoImpl) SelectFileByName(ctx context.Context, name string, storageID, fatherID int64) (files []*do.File, err error) {
	tx := f.db.WithContext(ctx).Where("storage_id = ?", storageID)
	if fatherID != 0 {
		tx = tx.Where("father_id = ?", fatherID)
	} else {
		tx = tx.Where("father_id = ? ", global.DefaultFatherID)
	}
	res := tx.Where("name = ?", name).Order("id").Limit(1).Find(&files)
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return files, res.Error
}

// SelectFileTypeCount 获取文件类型的统计数据
func (f *FileDaoImpl) SelectFileTypeCount(ctx context.Context, storageID int64) (types []*do.Statistics, err error) {
	err = f.db.WithContext(ctx).Debug().Model(&do.File{}).
		Select("type, count(*) as count").
		Where("storage_id = ? and deleted_at is null", storageID).
		Group("type").
		Scan(&types).Error
	return
}

var (
	fileDao     base.FileDao
	fileDaoOnce sync.Once
)

func GetFileDao(db *gorm.DB) base.FileDao {
	fileDaoOnce.Do(func() {
		fileDao = &FileDaoImpl{db: db}
	})
	return fileDao
}
