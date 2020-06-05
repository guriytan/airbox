package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type FileEntity struct {
	Model
	Hash string `gorm:"type:varchar(64)";json:"-"` // 文件Hash值
	Name string `gorm:"type:varchar(100);index"`   // 文件名
	Size uint64 // 文件大小
	Path string `gorm:"type:varchar(500)"` // 文件路径

	Link int // 文件关联次数
}

func (f *FileEntity) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}
