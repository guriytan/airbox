package do

import (
	"gorm.io/gorm"

	"airbox/utils"
)

// Storage 每个用户仅拥有一个数据仓库
type Storage struct {
	Model
	UserID      int64  `gorm:"type:bigint(20);index" json:"user_id,string"`         // 所属用户ID
	BucketName  string `gorm:"type:char(64)" json:"bucket_name"`                    // oss bucket name
	MaxSize     int64  `gorm:"type:bigint(20);default:21474836480" json:"max_size"` // 最大容量
	CurrentSize int64  `gorm:"type:bigint(20)" json:"current_size"`                 // 当前容量
}

func (s *Storage) BeforeCreate(tx *gorm.DB) error {
	if s.ID == 0 {
		s.ID = utils.GetSnowflake().Generate()
	}
	return nil
}
