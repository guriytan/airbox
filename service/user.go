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

func (u *UserService) Login(username, email, password string) (*model.User, error) {
	return u.user.SelectUserByPwdAndNameOrEmail(username, email, utils.EncryptPassword(password))
}

func (u *UserService) GetUserByID(id string) (*model.User, error) {
	return u.user.SelectUserByID(id)
}

func (u *UserService) GetUserByUsername(username string) (*model.User, error) {
	return u.user.SelectUserByName(username)
}

func (u *UserService) GetUserByEmail(email string) (*model.User, error) {
	return u.user.SelectUserByEmail(email)
}

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

func (u *UserService) ResetPwd(id, password string) error {
	return u.user.UpdateUser(&model.User{
		Model: model.Model{
			ID: id,
		},
		Password: utils.EncryptPassword(password),
	})
}

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

// 从缓存中读取key为email的值与code判断是否一致
// 当相等时返回true，不相等返回false
func (u *UserService) VerifyEmailCaptcha(email string, code string) bool {
	if captcha := u.redis.GetCaptcha(email); captcha == code {
		return true
	}
	return false
}

// 生成随机验证码发送至邮箱
func (u *UserService) SendCaptcha(email string) error {
	captcha := utils.GetEmailCaptcha()
	if err := u.redis.SetCaptcha(email, captcha); err != nil {
		return err
	}
	return utils.SendCaptcha(email, captcha)
}

// 根据邮箱生成链接发送至邮箱
func (u *UserService) SendResetLink(id, email string) error {
	captcha, err := utils.GenerateEmailToken(id)
	if err != nil {
		return err
	}
	return utils.SendResetLink(email, Env.Web.Site+"/#/reset/"+captcha)
}

// CloseUser 注销用户，删除数据仓库内所有文件以及文件夹
func (u *UserService) CloseUser(id, sid string) error {
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
	return os.RemoveAll(PrefixMasterDirectory + sid + "/")
}
