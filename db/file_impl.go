package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/global"
	"airbox/model/do"
	"airbox/model/dto"

	"gorm.io/gorm"
)

// FileDaoImpl 文件数据库操作实体
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

// SelectFileByName 在数据仓库下查看文件夹下是否有文件名为name的文件
func (f *FileDaoImpl) SelectFileByName(ctx context.Context, name string, storageID, fatherID int64) (files []*do.File, err error) {
	query := f.db.WithContext(ctx).Where("storage_id = ?", storageID)
	if fatherID != 0 {
		query = query.Where("father_id = ?", fatherID)
	} else {
		query = query.Where("father_id = ? ", global.DefaultFatherID)
	}
	res := query.Where("name = ?", name).Order("id").Limit(1).Find(&files)
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return files, res.Error
}

// SelectFileTypeCount 获取文件类型的统计数据
func (f *FileDaoImpl) SelectFileTypeCount(ctx context.Context, storageID int64) (types []*do.Statistics, err error) {
	err = f.db.WithContext(ctx).Model(&do.File{}).
		Select("type, count(*) as count").
		Where("storage_id = ? and type != ? and deleted_at is null", storageID, global.FileFolderType).
		Group("type").
		Scan(&types).Error
	return
}

// CountByCondition 根据查询条件获取总数
func (f *FileDaoImpl) CountByCondition(ctx context.Context, cond *dto.QueryCondition) (count int64, err error) {
	query := f.db.WithContext(ctx).Model(&do.File{}).Where("storage_id = ?", cond.StorageID)
	if cond.IsSetFatherID() {
		query = query.Where("father_id = ? ", cond.GetFatherID())
	}
	if cond.IsSetType() {
		query = query.Where("type = ? ", cond.GetType())
	}
	err = query.Count(&count).Error
	return
}

func (f *FileDaoImpl) SelectFileByCondition(ctx context.Context, cond *dto.QueryCondition) (files []*do.File, err error) {
	query := f.db.WithContext(ctx).Preload("FileInfo").Where("storage_id = ?", cond.StorageID)
	if cond.IsSetFatherID() {
		query = query.Where("father_id = ? ", cond.GetFatherID())
	}
	if cond.IsSetType() {
		query = query.Where("type = ? ", cond.GetType())
	}
	if cond.Cursor > 0 {
		query = query.Where("id < ?", cond.Cursor)
	}
	if cond.Limit > 0 {
		query = query.Limit(cond.Limit)
	}
	err = query.Order("id desc").Find(&files).Error
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
