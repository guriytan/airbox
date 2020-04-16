package service

import (
	"airbox/cache"
	. "airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/model"
	"airbox/utils"
	"os"
)

type UserService struct {
	file    base.FileDao
	folder  base.FolderDao
	user    base.UserDao
	redis   *cache.RedisClient
	storage base.StorageDao
}

var user *UserService

func GetUserService() *UserService {
	if user == nil {
		user = &UserService{
			file:    db.GetFileDao(),
			folder:  db.GetFolderDao(),
			user:    db.GetUserDao(),
			redis:   cache.GetRedisClient(),
			storage: db.GetStorageDao(),
		}
	}
	return user
}

// Login 验证用户名或邮箱以及密码是否正确
func (u *UserService) Login(user, password string) (*model.User, error) {
	return u.user.SelectUserByPwdAndNameOrEmail(user, utils.EncryptPassword(password))
}

// GetUserByID 由于从token解析得到的user信息并不是实时的，因此这里提供实时的获取用户信息供显示容量
func (u *UserService) GetUserByID(id string) (*model.User, error) {
	return u.user.SelectUserByID(id)
}

// GetUserByUsername 检验用户名是否重复
func (u *UserService) GetUserByUsername(username string) (*model.User, error) {
	return u.user.SelectUserByName(username)
}

// GetUserByEmail 检验邮箱是否重复
func (u *UserService) GetUserByEmail(email string) (*model.User, error) {
	return u.user.SelectUserByEmail(email)
}

// Registry 注册用户
func (u *UserService) Registry(username string, password string, email string) error {
	tx := DB.Begin()
	user := &model.User{
		Storage:  model.Storage{},
		Name:     username,
		Password: utils.EncryptPassword(password),
		Email:    email,
	}
	if err := u.user.InsertUser(tx, user); err != nil {
		tx.Rollback()
		return err
	}
	if err := u.storage.InsertStorage(tx, &model.Storage{UserID: user.ID}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	u.redis.DeleteCaptcha(email)
	return nil
}

// 修改Pwd 重置密码
func (u *UserService) ResetPwd(id, password string) error {
	return u.user.UpdateUser(&model.User{
		Model: model.Model{
			ID: id,
		},
		Password: utils.EncryptPassword(password),
	})
}

// ResetEmail 修改邮箱
func (u *UserService) ResetEmail(id, email string) error {
	if err := u.user.UpdateUser(&model.User{
		Model: model.Model{
			ID: id,
		},
		Email: email,
	}); err != nil {
		return err
	}
	u.redis.DeleteCaptcha(email)
	return nil
}

// UnsubscribeUser 注销用户，删除数据仓库内所有文件以及文件夹
func (u *UserService) UnsubscribeUser(id, sid string) error {
	tx := DB.Begin()
	if err := u.user.DeleteUserByID(id); err != nil {
		tx.Rollback()
		return err
	}
	if err := u.storage.DeleteStorageByID(tx, sid); err != nil {
		tx.Rollback()
		return err
	}
	if err := u.folder.DeleteFolderBySID(tx, sid); err != nil {
		tx.Rollback()
		return err
	}
	if err := u.file.DeleteFileBySID(tx, sid); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return os.RemoveAll(FilePrefixMasterDirectory + sid + "/")
}
