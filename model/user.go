package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Model

	Storage Storage // 对应数据仓库

	Name     string `gorm:"type:varchar(20);uniqueIndex"` // 用户名
	Password string `gorm:"type:varchar(80);index"`       // 密码
	Email    string `gorm:"type:varchar(50);uniqueIndex"` // 邮箱
}

func (user *User) BeforeCreate(tx *gorm.DB) error {
	if len(user.ID) == 0 {
		user.ID = uuid.New().String()
	}
	return nil
}
