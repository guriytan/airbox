package db

import (
	"context"
	"sync"

	"airbox/db/base"
	"airbox/model/do"

	"gorm.io/gorm"
)

// 用户数据库操作实体
type UserDaoImpl struct {
	db *gorm.DB
}

// InsertUser 新增用户
func (u *UserDaoImpl) InsertUser(ctx context.Context, tx *gorm.DB, user *do.User) error {
	if tx == nil {
		tx = u.db.WithContext(ctx)
	}
	return tx.Create(user).Error
}

// DeleteUserByID 根据用户ID删除用户
func (u *UserDaoImpl) DeleteUserByID(ctx context.Context, tx *gorm.DB, id string) error {
	if tx == nil {
		tx = u.db.WithContext(ctx)
	}
	return tx.Delete(&do.User{}, "id = ?", id).Error
}

// UpdateUser 更新用户信息
func (u *UserDaoImpl) UpdateUser(ctx context.Context, user *do.User) error {
	return u.db.WithContext(ctx).Model(&do.User{}).Updates(user).Error
}

// SelectUserByID 根据用户ID获得用户
func (u *UserDaoImpl) SelectUserByID(ctx context.Context, id string) (*do.User, error) {
	user := &do.User{}
	err := u.db.WithContext(ctx).Preload("Storage").Where("id = ?", id).Find(user).Error
	return user, err
}

// SelectUserByPwdAndNameOrEmail 根据用户名或邮箱以及密码获得用户
func (u *UserDaoImpl) SelectUserByPwdAndNameOrEmail(ctx context.Context, name, pwd string) (*do.User, error) {
	user := &do.User{}
	sql := u.db.WithContext(ctx).Preload("Storage")
	err := sql.Where("password = ? and (name = ? or email = ?)", pwd, name, name).Order("id").Limit(1).Find(user).Error
	return user, err
}

// SelectUserByName 根据用户名获得用户
func (u *UserDaoImpl) SelectUserByName(ctx context.Context, username string) (*do.User, error) {
	user := &do.User{}
	res := u.db.WithContext(ctx).Find(user, "name = ?", username).Error
	return user, res
}

// SelectUserByEmail 根据邮箱获得用户
func (u *UserDaoImpl) SelectUserByEmail(ctx context.Context, email string) (*do.User, error) {
	user := &do.User{}
	res := u.db.WithContext(ctx).Find(user, "email = ?", email).Error
	return user, res
}

var (
	userDao     base.UserDao
	userDaoOnce sync.Once
)

func GetUserDao(db *gorm.DB) base.UserDao {
	userDaoOnce.Do(func() {
		userDao = &UserDaoImpl{db: db}
	})
	return userDao
}
