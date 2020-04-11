package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// 每个用户仅拥有一个数据仓库
type Storage struct {
	Model
	UserID      string `gorm:"type:varchar(36)"`    // 所属用户ID
	MaxSize     uint64 `gorm:"default:21474836480"` // 最大容量
	CurrentSize uint64 // 当前容量
}

func (s *Storage) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}

func (Storage) TableName() string {
	return "storage"
}
