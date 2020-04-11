package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"time"
)

type Model struct {
	ID        string `gorm:"type:varchar(36);primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// 统计文件类型数量
type Statistics struct {
	Type  int
	Count int
}

type File struct {
	Model
	Hash      string  `gorm:"type:varchar(64)"`       // 文件Hash值
	Name      string  `gorm:"type:varchar(50);index"` // 文件名
	Size      uint64  // 文件大小
	Location  string  `gorm:"type:varchar(200)"`      // 文件路径
	FolderID  *string `gorm:"type:varchar(36);index"` // 所在文件夹ID，当ID为nil时文件直属数据仓库下
	StorageID string  `gorm:"type:varchar(36);index"` // 所在数据仓库ID
	Suffix    string  `gorm:"type:varchar(20)"`       // 文件后缀
	Type      int     `gorm:"index"`                  // 文件类型
}

func (f *File) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}

func (File) TableName() string {
	return "file"
}
