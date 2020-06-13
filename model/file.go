package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"time"
)

type Model struct {
	ID        string     `gorm:"type:varchar(36);primary_key"`
	CreatedAt time.Time  `gorm:"index"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index";json:"-"`
}

// 统计文件类型数量
type Statistics struct {
	Type  int
	Count int
}

type File struct {
	Model
	Name      string  `gorm:"type:varchar(100);index"` // 文件名
	FolderID  *string `gorm:"type:varchar(36);index"`  // 所在文件夹ID，当ID为nil时文件直属数据仓库下
	StorageID string  `gorm:"type:varchar(36);index"`  // 所在数据仓库ID
	Type      int     `gorm:"index"`                   // 文件类型

	FileInfoID string `gorm:"type:varchar(36);index"`
	FileInfo   FileInfo
}

func (f *File) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}

type FileInfo struct {
	Model
	Hash string `gorm:"type:varchar(64)";json:"-"` // 文件Hash值
	Name string `gorm:"type:varchar(100);index"`   // 文件名
	Size uint64 // 文件大小
	Path string `gorm:"type:varchar(500)"` // 文件路径

	FileCount FileCount
}

func (f *FileInfo) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}

type FileCount struct {
	Model
	FileInfoID string `gorm:"type:varchar(64);index"` // 文件Hash值
	Link       int    // 文件关联次数
}

func (f *FileCount) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}
