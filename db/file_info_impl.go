package db

import (
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 文件数据库操作实体
type FileInfoDaoImpl struct {
}

func (f *FileInfoDaoImpl) DeleteFileInfo(db *gorm.DB, id string) error {
	return db.Delete(&model.FileInfo{}, "id = ?", id).Error
}

func (f *FileInfoDaoImpl) DeleteFileCount(db *gorm.DB, id string) error {
	return db.Delete(&model.FileCount{}, "file_info_id = ?", id).Error
}

func (f *FileInfoDaoImpl) InsertFileCount(db *gorm.DB, count *model.FileCount) error {
	return db.Create(count).Error
}

func (f *FileInfoDaoImpl) SelectFileCountByInfoID(db *gorm.DB, id string) (*model.FileCount, error) {
	count := &model.FileCount{}
	err := db.Where("file_info_id = ?", id).First(count).Error
	return count, err
}

func (f *FileInfoDaoImpl) SelectFileInfoByID(db *gorm.DB, id string) (*model.FileInfo, error) {
	info := &model.FileInfo{}
	err := db.Where("id = ?", id).First(info).Error
	return info, err
}

func (f *FileInfoDaoImpl) SelectFileInfoByMD5(db *gorm.DB, md5 string) (*model.FileInfo, error) {
	info := &model.FileInfo{}
	err := db.Where("hash = ?", md5).First(info).Error
	return info, err
}

func (f *FileInfoDaoImpl) UpdateFileInfo(db *gorm.DB, id string, delta int64) error {
	return db.Table("file_count").Where("file_info_id = ?", id).UpdateColumn("link", gorm.Expr("link + ?", delta)).Error
}

func (f *FileInfoDaoImpl) InsertFileInfo(db *gorm.DB, info *model.FileInfo) error {
	return db.Create(info).Error
}

var infoDao base.FileInfoDao

func GetFileInfoDao() base.FileInfoDao {
	if infoDao == nil {
		infoDao = &FileInfoDaoImpl{}
	}
	return infoDao
}
