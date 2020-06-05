package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 用户数据库操作接口
type UserDao interface {
	InsertUser(db *gorm.DB, user *model.User) error

	DeleteUserByID(db *gorm.DB, id string) error

	UpdateUser(db *gorm.DB, user *model.User) error

	SelectUserByID(db *gorm.DB, id string) (*model.User, error)
	SelectUserByPwdAndNameOrEmail(db *gorm.DB, name, pwd string) (*model.User, error)
	SelectUserByName(db *gorm.DB, username string) (*model.User, error)
	SelectUserByEmail(db *gorm.DB, email string) (*model.User, error)
}
