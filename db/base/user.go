package base

import (
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 用户数据库操作接口
type UserDao interface {
	InsertUser(db *gorm.DB, user *model.User) error

	DeleteUserByID(id string) error

	UpdateUser(user *model.User) error

	SelectUserByID(id string) (*model.User, error)
	SelectUserByPwdAndNameOrEmail(name, pwd string) (*model.User, error)
	SelectUserByName(username string) (*model.User, error)
	SelectUserByEmail(email string) (*model.User, error)
}
