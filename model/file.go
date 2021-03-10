package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        string `gorm:"type:varchar(36);primary_key"`
	CreatedAt int64  `gorm:"index"`
	UpdatedAt int64
	DeletedAt int64 `gorm:"index"`
}

// 统计文件类型数量
type Statistics struct {
	Type  int
	Count int
}

type File struct {
	Model
	StorageID string `gorm:"type:varchar(36);index:idx_file"`  // 所在数据仓库ID
	FatherID  string `gorm:"type:varchar(36);index:idx_file"`  // 所在文件夹ID，当ID为nil时文件直属数据仓库下
	Name      string `gorm:"type:varchar(512);index:idx_file"` // 文件名
	Type      int    `gorm:"index:idx_type"`                   // 文件类型

	FileInfoID string `gorm:"type:varchar(36);index"`
	FileInfo   FileInfo
}

func (f *File) BeforeCreate(tx *gorm.DB) error {
	if len(f.ID) == 0 {
		f.ID = uuid.New().String()
	}
	return nil
}

type FileInfo struct {
	Model
	Hash string `gorm:"type:varchar(64);index:idx_hash,unique"` // 文件Hash值
	Name string `gorm:"type:varchar(512)"`                      // 文件名
	Size uint64 // 文件大小
	Path string `gorm:"type:text"` // 文件路径
	Link int    // 文件关联次数
}

func (f *FileInfo) BeforeCreate(tx *gorm.DB) error {
	if len(f.ID) == 0 {
		f.ID = uuid.New().String()
	}
	return nil
}
