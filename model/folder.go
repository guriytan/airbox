package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Folder struct {
	Model
	StorageID string  `gorm:"type:varchar(36);index"` // 所处数据仓库ID
	FatherID  *string `gorm:"type:varchar(36);index"` // 父文件夹ID，当ID为0时文件夹直属数据仓库
	Name      string  `gorm:"type:varchar(50);index"` //文件夹名
}

func (f *Folder) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}

func (Folder) TableName() string {
	return "folder"
}
