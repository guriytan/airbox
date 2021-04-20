package do

import (
	"time"

	"gorm.io/gorm"

	"airbox/utils"
)

type Model struct {
	ID        int64      `gorm:"type:bigint(20);primary_key" json:"id,string"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
}

// Statistics 统计文件类型数量
type Statistics struct {
	Type  int `json:"type"`
	Count int `json:"count"`
}

type File struct {
	Model
	StorageID int64  `gorm:"type:bigint(20);index:idx_file" json:"storage_id,string"` // 所在数据仓库ID
	FatherID  int64  `gorm:"type:bigint(20);index:idx_file" json:"father_id,string"`  // 所在文件夹ID，当ID为nil时文件直属数据仓库下
	Name      string `gorm:"type:varchar(512);index:idx_file" json:"name"`            // 文件名
	Type      int    `gorm:"index:idx_type" json:"type"`                              // 文件类型

	FileInfoID int64    `gorm:"type:bigint(20);index" json:"file_info_id,string"`
	FileInfo   FileInfo `json:"file_info"`
}

func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == 0 {
		f.ID = utils.GetSnowflake().Generate()
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
	if f.ID == 0 {
		f.ID = utils.GetSnowflake().Generate()
	}
	return nil
}
