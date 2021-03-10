package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/global"
	"airbox/model"

	"gorm.io/gorm"
)

// 文件数据库操作实体
type FileDaoImpl struct {
	db *gorm.DB
}

// InsertFile 新增文件
func (f *FileDaoImpl) InsertFile(ctx context.Context, tx *gorm.DB, file *model.File) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Create(file).Error
}

// DeleteFileByID 根据文件ID删除文件
func (f *FileDaoImpl) DeleteFileByID(ctx context.Context, id string) error {
	return f.db.WithContext(ctx).Delete(&model.File{}, "id = ?", id).Error
}

// DeleteFileByStorageID 根据仓库ID删除文件
func (f *FileDaoImpl) DeleteFileByStorageID(ctx context.Context, tx *gorm.DB, storageID string) error {
	if tx == nil {
		tx = f.db.WithContext(ctx)
	}
	return tx.Delete(&model.File{}, "storage_id = ?", storageID).Error
}

// UpdateFile 更新文件信息
func (f *FileDaoImpl) UpdateFile(ctx context.Context, id string, file map[string]interface{}) error {
	return f.db.WithContext(ctx).Model(&model.File{}).Where("id = ?", id).Updates(file).Error
}

// SelectFileByID 根据文件ID获得文件
func (f *FileDaoImpl) SelectFileByID(ctx context.Context, id string) (*model.File, error) {
	file := &model.File{}
	err := f.db.WithContext(ctx).Preload("FileInfo").Where("id = ?", id).Order("id").Limit(1).Find(file).Error
	return file, err
}

// SelectFileByStorageID 获取在数据仓库Sid下的文件
func (f *FileDaoImpl) SelectFileByStorageID(ctx context.Context, storageID string) (files []*model.File, err error) {
	err = f.db.WithContext(ctx).Preload("FileInfo").Where("storage_id = ? and father_id = ?", storageID, global.DefaultFatherID).Order("created_at desc").Find(&files).Error
	return
}

// SelectFileByFatherID 根据文件夹ID获得文件
func (f *FileDaoImpl) SelectFileByFatherID(ctx context.Context, id string) (files []*model.File, err error) {
	err = f.db.WithContext(ctx).Preload("FileInfo").Where("father_id = ?", id).Order("created_at desc").Find(&files).Error
	return
}

// SelectFileByType 根据文件类型获得文件
func (f *FileDaoImpl) SelectFileByType(ctx context.Context, fileType string) (files []*model.File, err error) {
	err = f.db.WithContext(ctx).Preload("FileInfo").Where("type = ?", fileType).Order("created_at desc").Find(&files).Error
	return
}

// SelectFileByName 在数据仓库下查看文件夹下是否有文件名为name的文件
func (f *FileDaoImpl) SelectFileByName(ctx context.Context, name, storageID, fatherID string) (*model.File, error) {
	tx := f.db.WithContext(ctx).Where("storage_id = ?", storageID)
	if len(fatherID) != 0 {
		tx = tx.Where("father_id = ?", fatherID)
	} else {
		tx = tx.Where("father_id = ? ", global.DefaultFatherID)
	}
	file := &model.File{}
	if err := tx.Where("name = ?", name).Order("id").Limit(1).Find(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// SelectFileTypeCount 获取文件类型的统计数据
func (f *FileDaoImpl) SelectFileTypeCount(ctx context.Context, storageID string) (types []*model.Statistics, err error) {
	err = f.db.WithContext(ctx).Model(&model.File{}).Select("type, count(*) as count").Where("deleted_at = 0 and storage_id = ?", storageID).
		Group("type").Scan(&types).Error
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
