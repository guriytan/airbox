package do

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"airbox/utils/encryption"
)

var (
	ErrNotSupportUpdateParam = errors.New("update need to use user struct")
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
	if len(user.Password) != 0 {
		user.Password = encryption.EncryptPassword(user.Password)
	}
	return nil
}

func (user *User) BeforeUpdate(tx *gorm.DB) error {
	const password = "password"
	if tx.Statement.Changed(password) {
		raw, ok := tx.Statement.Dest.(*User)
		if !ok {
			return ErrNotSupportUpdateParam
		}
		tx.Statement.SetColumn(password, encryption.EncryptPassword(raw.Password))
	}
	return nil
}
