package service

import (
	"context"
	"os"
	"sync"

	"airbox/config"
	"airbox/db"
	"airbox/db/base"
	"airbox/logger"
	"airbox/model"
	"airbox/utils/encryption"

	"gorm.io/gorm"
)

type UserService struct {
	file    base.FileDao
	user    base.UserDao
	storage base.StorageDao
}

var (
	userService     *UserService
	userServiceOnce sync.Once
)

func GetUserService() *UserService {
	userServiceOnce.Do(func() {
		userService = &UserService{
			file:    db.GetFileDao(config.GetDB()),
			user:    db.GetUserDao(config.GetDB()),
			storage: db.GetStorageDao(config.GetDB()),
		}
	})
	return userService
}

// Login 验证用户名或邮箱以及密码是否正确
func (u *UserService) Login(ctx context.Context, user, password string) (*model.User, error) {
	byPwdAndNameOrEmail, err := u.user.SelectUserByPwdAndNameOrEmail(ctx, user, encryption.EncryptPassword(password))
	if err != nil {
		return nil, err
	}
	return byPwdAndNameOrEmail, nil
}

// GetUserByID 由于从token解析得到的user信息并不是实时的，因此这里提供实时的获取用户信息供显示容量
func (u *UserService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	byID, err := u.user.SelectUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return byID, nil
}

// GetUserByUsername 检验用户名是否重复
func (u *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, bool) {
	log := logger.GetLogger(ctx, "GetUserByUsername")
	byName, err := u.user.SelectUserByName(ctx, username)
	if err == gorm.ErrRecordNotFound {
		return byName, true
	} else if err != nil {
		log.Infof("%s\n", err.Error())
	}
	return byName, false
}

// GetUserByEmail 检验邮箱是否重复
func (u *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, bool) {
	log := logger.GetLogger(ctx, "GetUserByUsername")
	byEmail, err := u.user.SelectUserByEmail(ctx, email)
	if err == gorm.ErrRecordNotFound {
		return byEmail, true
	} else if err != nil {
		log.Infof("%s\n", err.Error())
	}
	return byEmail, false
}

// Registry 注册用户
func (u *UserService) Registry(ctx context.Context, username string, password string, email string) error {
	user := &model.User{
		Storage:  model.Storage{},
		Name:     username,
		Password: encryption.EncryptPassword(password),
		Email:    email,
	}
	storage := &model.Storage{UserID: user.ID}
	err := config.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := u.user.InsertUser(ctx, tx, user); err != nil {
			return err
		}
		if err := u.storage.InsertStorage(ctx, tx, storage); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// 修改Pwd 重置密码
func (u *UserService) ResetPwd(ctx context.Context, id, password string) error {
	if err := u.user.UpdateUser(ctx, &model.User{
		Model: model.Model{
			ID: id,
		},
		Password: encryption.EncryptPassword(password),
	}); err != nil {
		return err
	}
	return nil
}

// ResetEmail 修改邮箱
func (u *UserService) ResetEmail(ctx context.Context, id, email string) error {
	if err := u.user.UpdateUser(ctx, &model.User{
		Model: model.Model{
			ID: id,
		},
		Email: email,
	}); err != nil {
		return err
	}
	return nil
}

// UnsubscribeUser 注销用户，删除数据仓库内所有文件以及文件夹
func (u *UserService) UnsubscribeUser(ctx context.Context, id, sid string) error {
	err := config.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := u.user.DeleteUserByID(ctx, tx, id); err != nil {
			return err
		}
		if err := u.storage.DeleteStorageByID(ctx, tx, sid); err != nil {
			return err
		}
		if err := u.file.DeleteFileByStorageID(ctx, tx, sid); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := os.RemoveAll(config.GetConfig().Upload.Dir + "/" + sid + "/"); err != nil {
		return err
	}
	return nil
}
