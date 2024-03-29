package do

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"airbox/utils"
	"airbox/utils/encryption"
	"airbox/utils/hasher"
)

var (
	ErrNotSupportUpdateParam = errors.New("update need to use user struct")
)

type User struct {
	Model
	Name     string `gorm:"type:varchar(20);uniqueIndex" json:"name"`  // 用户名
	Password string `gorm:"type:varchar(80);index" json:"password"`    // 密码
	Hash     string `gorm:"type:char(64);index" json:"hash"`           // 密码Hash
	Email    string `gorm:"type:varchar(50);uniqueIndex" json:"email"` // 邮箱

	Storage Storage `json:"storage"` // 对应数据仓库
}

func (user *User) BeforeCreate(tx *gorm.DB) error {
	if user.ID == 0 {
		user.ID = utils.GetSnowflake().Generate()
	}
	if len(user.Password) != 0 {
		user.Hash = hasher.GetSha256().Hash(user.Password)
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
