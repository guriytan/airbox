package db

import (
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 用户数据库操作实体
type UserDaoImpl struct {
}

// InsertUser 新增用户
func (u *UserDaoImpl) InsertUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// DeleteUserByID 根据用户ID删除用户
func (u *UserDaoImpl) DeleteUserByID(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.User{}).Error
}

// UpdateUser 更新用户信息
func (u *UserDaoImpl) UpdateUser(db *gorm.DB, user *model.User) error {
	return db.Model(user).Updates(user).Error
}

// SelectUserByID 根据用户ID获得用户
func (u *UserDaoImpl) SelectUserByID(db *gorm.DB, id string) (*model.User, error) {
	user := &model.User{}
	err := db.Preload("Storage").Where("id = ?", id).First(user).Error
	return user, err
}

// SelectUserByPwdAndNameOrEmail 根据用户名或邮箱以及密码获得用户
func (u *UserDaoImpl) SelectUserByPwdAndNameOrEmail(db *gorm.DB, name, pwd string) (*model.User, error) {
	user := &model.User{}
	sql := db.Preload("Storage")
	err := sql.Where("password = ? and (name = ? or email = ?)", pwd, name, name).First(user).Error
	return user, err
}

// SelectUserByName 根据用户名获得用户
func (u *UserDaoImpl) SelectUserByName(db *gorm.DB, username string) (*model.User, error) {
	user := &model.User{}
	res := db.Where("name = ?", username).First(user).Error
	return user, res
}

// SelectUserByEmail 根据邮箱获得用户
func (u *UserDaoImpl) SelectUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	user := &model.User{}
	res := db.Where("email = ?", email).First(user).Error
	return user, res
}

var user base.UserDao

func GetUserDao() base.UserDao {
	if user == nil {
		user = &UserDaoImpl{}
	}
	return user
}
