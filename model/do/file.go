package do

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        string     `gorm:"type:varchar(36);primary_key" json:"id"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
}

// 统计文件类型数量
type Statistics struct {
	Type  int `json:"type"`
	Count int `json:"count"`
}

type File struct {
	Model
	StorageID string `gorm:"type:char(36);index:idx_file" json:"storage_id"` // 所在数据仓库ID
	FatherID  string `gorm:"type:char(36);index:idx_file" json:"father_id"`  // 所在文件夹ID，当ID为nil时文件直属数据仓库下
	Name      string `gorm:"type:varchar(512);index:idx_file" json:"name"`   // 文件名
	Type      int    `gorm:"index:idx_type" json:"type"`                     // 文件类型

	FileInfoID string   `gorm:"type:varchar(36);index" json:"file_info_id"`
	FileInfo   FileInfo `json:"file_info"`
}

func (f *File) BeforeCreate(tx *gorm.DB) error {
	if len(f.ID) == 0 {
		f.ID = uuid.New().String()
	}
	return nil
}

type FileInfo struct {
	Model
	Hash   string `gorm:"type:char(64);index:idx_hash,unique" json:"hash"` // 文件Hash值
	Name   string `gorm:"type:varchar(512)" json:"name"`                   // 文件名
	OssKey string `gorm:"type:char(64)" json:"oss_key"`                    // Oss key
	Size   int64  `gorm:"type:bigint(20)" json:"size"`                     // 文件大小
	Link   int    `gorm:"type:int(11)" json:"link"`                        // 文件关联次数
}

func (f *FileInfo) BeforeCreate(tx *gorm.DB) error {
	if len(f.ID) == 0 {
		f.ID = uuid.New().String()
	}
	return nil
}
