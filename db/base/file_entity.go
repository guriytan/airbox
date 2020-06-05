package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

type FileEntityDao interface {
	InsertFileEntity(db *gorm.DB, entity *model.FileEntity) error
	UpdateFileEntity(db *gorm.DB, id string, size int64) error
	SelectFileEntityByID(db *gorm.DB, id string) (*model.FileEntity, error)
	SelectFileEntityByMD5(db *gorm.DB, md5 string) (*model.FileEntity, error)
}
