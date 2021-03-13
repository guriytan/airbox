package do

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 每个用户仅拥有一个数据仓库
type Storage struct {
	Model
	UserID      string `gorm:"type:char(36);index"`                 // 所属用户ID
	BucketName  string `gorm:"type:char(64)"`                       // oss bucket name
	MaxSize     int64  `gorm:"type:bigint(20);default:21474836480"` // 最大容量
	CurrentSize int64  `gorm:"type:bigint(20)"`                     // 当前容量
}

func (s *Storage) BeforeCreate(tx *gorm.DB) error {
	if len(s.ID) == 0 {
		s.ID = uuid.New().String()
	}
	return nil
}
