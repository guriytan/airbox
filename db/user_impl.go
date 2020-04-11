package db

import (
	. "airbox/config"
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 用户数据库操作实体
type UserDaoImpl struct {
}

// 新增用户
func (u *UserDaoImpl) InsertUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// 根据用户ID删除用户
func (u *UserDaoImpl) DeleteUserByID(id string) error {
	return DB.Where("id = ?", id).Delete(&model.User{}).Error
}

// 更新用户信息
func (u *UserDaoImpl) UpdateUser(user *model.User) error {
	return DB.Model(user).Updates(user).Error
}

// 根据用户ID获得用户
func (u *UserDaoImpl) SelectUserByID(id string) (*model.User, error) {
	user := &model.User{}
	err := DB.Preload("Storage").Where("id = ?", id).First(user).Error
	return user, err
}

// 根据用户名或邮箱以及密码获得用户
func (u *UserDaoImpl) SelectUserByPwdAndNameOrEmail(name, email, pwd string) (*model.User, error) {
	user := &model.User{}
	db := DB.Preload("Storage")
	if name != "" {
		db = db.Where("name = ?", name)
	} else if email != "" {
		db = db.Where("email = ?", email)
	}
	err := db.Where("password = ?", pwd).First(user).Error
	return user, err
}

// 根据用户名获得用户
func (u *UserDaoImpl) SelectUserByName(username string) (*model.User, error) {
	user := &model.User{}
	err := DB.Where("name = ?", username).First(user).Error
	return user, err
}

// 根据邮箱获得用户
func (u *UserDaoImpl) SelectUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := DB.Where("email = ?", email).First(user).Error
	return user, err
}

var user base.UserDao

func GetUserDao() base.UserDao {
	if user == nil {
		user = &UserDaoImpl{}
	}
	return user
}
