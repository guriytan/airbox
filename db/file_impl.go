package db

import (
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 文件数据库操作实体
type FileDaoImpl struct {
}

func (f *FileDaoImpl) SelectFileTypeCount(db *gorm.DB) (types []model.Statistics, err error) {
	err = db.Table("file").Select("type, count(*) as count").Where("deleted_at is null").
		Group("type").Scan(&types).Error
	return
}

// 根据文件ID删除文件
func (f *FileDaoImpl) DeleteFileByID(db *gorm.DB, id string) error {
	return db.Delete(&model.File{}, "id = ?", id).Error
}

// 根据仓库ID删除文件
func (f *FileDaoImpl) DeleteFileBySID(db *gorm.DB, sid string) error {
	return db.Delete(&model.File{}, "storage_id = ?", sid).Error
}

// 在数据仓库下查看文件夹下是否有文件名为name的文件
func (f *FileDaoImpl) SelectFileByName(db *gorm.DB, name, sid string, fid *string) (*model.File, error) {
	tx := db.Where("storage_id = ?", sid)
	if fid != nil {
		tx = tx.Where("folder_id = ?", fid)
	} else {
		tx = tx.Where("folder_id is null")
	}
	file := &model.File{}
	err := tx.Where("name = ?", name).First(file).Error
	return file, err
}

// 获取在数据仓库Sid下，父文件夹为Fid的文件
func (f *FileDaoImpl) SelectFileByStorageID(db *gorm.DB, sid string) (files []model.File, err error) {
	err = db.Where("storage_id = ? and folder_id is null", sid).Find(&files).Error
	return
}

// 新增文件
func (f *FileDaoImpl) InsertFile(db *gorm.DB, file *model.File) error {
	return db.Create(file).Error
}

// 更新文件信息
func (f *FileDaoImpl) UpdateFile(db *gorm.DB, file *model.File) error {
	return db.Model(file).Update("name", file.Name).Error
}

// 根据文件ID获得文件
func (f *FileDaoImpl) SelectFileByID(db *gorm.DB, id string) (*model.File, error) {
	file := &model.File{}
	err := db.Where("id = ?", id).First(file).Error
	return file, err
}

// 根据文件夹ID获得文件
func (f *FileDaoImpl) SelectFileByFolderID(db *gorm.DB, id string) (files []model.File, err error) {
	err = db.Where("folder_id = ?", id).Find(&files).Error
	return
}

// 根据文件类型获得文件
func (f *FileDaoImpl) SelectFileByType(db *gorm.DB, t string) (files []model.File, err error) {
	err = db.Where("type = ?", t).Find(&files).Error
	return
}

var file base.FileDao

func GetFileDao() base.FileDao {
	if file == nil {
		file = &FileDaoImpl{}
	}
	return file
}
