package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type User struct {
	Model

	Storage Storage // 对应数据仓库

	Name     string `gorm:"type:varchar(20);index"` // 用户名
	Password string `gorm:"type:varchar(80);index"` // 密码
	Email    string `gorm:"type:varchar(50);index"` // 邮箱
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New().String())
}

func (User) TableName() string {
	return "user"
}
