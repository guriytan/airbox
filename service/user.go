package service

import (
	"airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/global"
	"airbox/model"
	"airbox/utils/encryption"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"os"
)

type UserService struct {
	file    base.FileDao
	folder  base.FolderDao
	user    base.UserDao
	storage base.StorageDao
}

var user *UserService

func GetUserService() *UserService {
	if user == nil {
		user = &UserService{
			file:    db.GetFileDao(),
			folder:  db.GetFolderDao(),
			user:    db.GetUserDao(),
			storage: db.GetStorageDao(),
		}
	}
	return user
}

// Login 验证用户名或邮箱以及密码是否正确
func (u *UserService) Login(user, password string) (*model.User, error) {
	byPwdAndNameOrEmail, err := u.user.SelectUserByPwdAndNameOrEmail(global.DB, user, encryption.EncryptPassword(password))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byPwdAndNameOrEmail, nil
}

// GetUserByID 由于从token解析得到的user信息并不是实时的，因此这里提供实时的获取用户信息供显示容量
func (u *UserService) GetUserByID(id string) (*model.User, error) {
	byID, err := u.user.SelectUserByID(global.DB, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return byID, nil
}

// GetUserByUsername 检验用户名是否重复
func (u *UserService) GetUserByUsername(username string) (*model.User, bool) {
	byName, err := u.user.SelectUserByName(global.DB, username)
	if err == gorm.ErrRecordNotFound {
		return byName, true
	} else if err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
	}
	return byName, false
}

// GetUserByEmail 检验邮箱是否重复
func (u *UserService) GetUserByEmail(email string) (*model.User, bool) {
	byEmail, err := u.user.SelectUserByEmail(global.DB, email)
	if err == gorm.ErrRecordNotFound {
		return byEmail, true
	} else if err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
	}
	return byEmail, false
}

// Registry 注册用户
func (u *UserService) Registry(username string, password string, email string) error {
	tx := global.DB.Begin()
	user := &model.User{
		Storage:  model.Storage{},
		Name:     username,
		Password: encryption.EncryptPassword(password),
		Email:    email,
	}
	if err := u.user.InsertUser(tx, user); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := u.storage.InsertStorage(tx, &model.Storage{UserID: user.ID}); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// 修改Pwd 重置密码
func (u *UserService) ResetPwd(id, password string) error {
	if err := u.user.UpdateUser(global.DB, &model.User{
		Model: model.Model{
			ID: id,
		},
		Password: encryption.EncryptPassword(password),
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ResetEmail 修改邮箱
func (u *UserService) ResetEmail(id, email string) error {
	if err := u.user.UpdateUser(global.DB, &model.User{
		Model: model.Model{
			ID: id,
		},
		Email: email,
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UnsubscribeUser 注销用户，删除数据仓库内所有文件以及文件夹
func (u *UserService) UnsubscribeUser(id, sid string) error {
	tx := global.DB.Begin()
	if err := u.user.DeleteUserByID(tx, id); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := u.storage.DeleteStorageByID(tx, sid); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := u.folder.DeleteFolderBySID(tx, sid); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := u.file.DeleteFileBySID(tx, sid); err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}
	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}
	err := os.RemoveAll(config.Env.Upload.Dir + "/" + sid + "/")
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
