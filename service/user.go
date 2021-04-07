package service

import (
	"context"
	"fmt"
	"sync"

	"airbox/db"
	"airbox/db/base"
	"airbox/logger"
	"airbox/model/do"
	"airbox/pkg"
	"airbox/utils/hasher"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
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
			file:    db.GetFileDao(pkg.GetDB()),
			user:    db.GetUserDao(pkg.GetDB()),
			storage: db.GetStorageDao(pkg.GetDB()),
		}
	})
	return userService
}

// Login 验证用户名或邮箱以及密码是否正确
func (u *UserService) Login(ctx context.Context, user, password string) (*do.User, error) {
	log := logger.GetLogger(ctx, "Login")
	byPwdAndNameOrEmail, err := u.user.SelectUserByPwdAndNameOrEmail(ctx, user, password)
	if err != nil {
		log.WithError(err).Warnf("get user by user key: %v failed", user)
		return nil, err
	}
	return byPwdAndNameOrEmail, nil
}

// GetUserByID 由于从token解析得到的user信息并不是实时的，因此这里提供实时的获取用户信息供显示容量
func (u *UserService) GetUserByID(ctx context.Context, userID int64) (*do.User, error) {
	log := logger.GetLogger(ctx, "GetUserByID")
	byID, err := u.user.SelectUserByID(ctx, userID)
	if err != nil {
		log.WithError(err).Warnf("get user info by id: %v failed", userID)
		return nil, err
	}
	if byID.Storage.CurrentSize > byID.Storage.MaxSize {
		byID.Storage.CurrentSize = byID.Storage.MaxSize
	}
	return byID, nil
}

// GetUserByUsername 检验用户名是否重复
func (u *UserService) GetUserByUsername(ctx context.Context, username string) (*do.User, bool) {
	log := logger.GetLogger(ctx, "GetUserByUsername")
	byName, err := u.user.SelectUserByName(ctx, username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	} else if err != nil {
		log.WithError(err).Warnf("get user by username: %v failed", username)
		return nil, false
	}
	return byName, true
}

// GetUserByEmail 检验邮箱是否重复
func (u *UserService) GetUserByEmail(ctx context.Context, email string) (*do.User, bool) {
	log := logger.GetLogger(ctx, "GetUserByEmail")
	byEmail, err := u.user.SelectUserByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	} else if err != nil {
		log.WithError(err).Warnf("get user by email: %v failed", email)
		return nil, false
	}
	return byEmail, true
}

// Registry 注册用户
func (u *UserService) Registry(ctx context.Context, username string, password string, email string) error {
	log := logger.GetLogger(ctx, "Registry")
	user := &do.User{
		Storage:  do.Storage{},
		Name:     username,
		Password: password,
		Email:    email,
	}
	err := pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := u.user.InsertUser(ctx, tx, user); err != nil {
			log.WithError(err).Warnf("save user: %+v failed", user)
			return err
		}
		storage := &do.Storage{UserID: user.ID, BucketName: fmt.Sprintf("bucket-%v", hasher.GetMD5().Hash(user.ID, username))}
		if err := u.storage.InsertStorage(ctx, tx, storage); err != nil {
			log.WithError(err).Warnf("save storage: %+v failed", storage)
			return err
		}
		if err := pkg.GetOSS().MakeBucket(ctx, storage.BucketName, minio.MakeBucketOptions{}); err != nil {
			log.WithError(err).Warnf("make bucket is oss: %v failed", storage.BucketName)
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Warn("transaction failed")
		return err
	}
	log.Infof("registry user: %+v success", user)
	return nil
}

// 修改Pwd 重置密码
func (u *UserService) ResetPwd(ctx context.Context, userID int64, password string) error {
	log := logger.GetLogger(ctx, "ResetPwd")
	if err := u.user.UpdateUser(ctx, &do.User{
		Model:    do.Model{ID: userID},
		Password: password,
	}); err != nil {
		log.WithError(err).Warnf("update user: %v password: %v failed", userID, password)
		return err
	}
	return nil
}

// ResetEmail 修改邮箱
func (u *UserService) ResetEmail(ctx context.Context, userID int64, email string) error {
	log := logger.GetLogger(ctx, "ResetEmail")
	if err := u.user.UpdateUser(ctx, &do.User{
		Model: do.Model{ID: userID},
		Email: email,
	}); err != nil {
		log.WithError(err).Warnf("update user: %v email: %v failed", userID, email)
		return err
	}
	return nil
}

// UnsubscribeUser 注销用户，删除数据仓库内所有文件以及文件夹
func (u *UserService) UnsubscribeUser(ctx context.Context, userID, storageID int64) error {
	log := logger.GetLogger(ctx, "UnsubscribeUser")
	err := pkg.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := u.user.DeleteUserByID(ctx, tx, userID); err != nil {
			log.WithError(err).Warnf("delete user: %v failed", userID)
			return err
		}
		if err := u.storage.DeleteStorageByID(ctx, tx, storageID); err != nil {
			log.WithError(err).Warnf("delete storage: %v failed", storageID)
			return err
		}
		if err := u.file.DeleteFileByStorageID(ctx, tx, storageID); err != nil {
			log.WithError(err).Warnf("delete filed of storage: %v failed", storageID)
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Warn("transaction failed")
		return err
	}
	log.Infof("unsubscribe user: %v success", userID)
	return nil
}
