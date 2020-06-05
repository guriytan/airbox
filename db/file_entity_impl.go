package db

import (
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 文件数据库操作实体
type FileEntityDaoImpl struct {
}

func (f *FileEntityDaoImpl) SelectFileEntityByID(db *gorm.DB, id string) (*model.FileEntity, error) {
	entity := &model.FileEntity{}
	err := db.Where("id = ?", id).First(entity).Error
	return entity, err
}

func (f *FileEntityDaoImpl) SelectFileEntityByMD5(db *gorm.DB, md5 string) (*model.FileEntity, error) {
	entity := &model.FileEntity{}
	err := db.Where("hash = ?", md5).First(entity).Error
	return entity, err
}

func (f *FileEntityDaoImpl) UpdateFileEntity(db *gorm.DB, id string, delta int64) error {
	return db.Table("file_entity").Where("id = ?", id).UpdateColumn("link", gorm.Expr("link + ?", delta)).Error
}

func (f *FileEntityDaoImpl) InsertFileEntity(db *gorm.DB, entity *model.FileEntity) error {
	return db.Create(entity).Error
}

var fileEntity base.FileEntityDao

func GetFileEntityDao() base.FileEntityDao {
	if fileEntity == nil {
		fileEntity = &FileEntityDaoImpl{}
	}
	return fileEntity
}
