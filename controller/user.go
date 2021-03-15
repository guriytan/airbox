package controller

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"airbox/config"
	"airbox/global"
	"airbox/logger"
	"airbox/model/vo"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"
)

// UserController is responsible for the request of user operation
type UserController struct {
	BaseController
	user   *service.UserService
	verify *service.AuthService
}

var (
	user     *UserController
	userOnce sync.Once
)

// GetUserController return instance of UserController
func GetUserController() *UserController {
	userOnce.Do(func() {
		user = &UserController{
			BaseController: getBaseController(),
			user:           service.GetUserService(),
			verify:         service.GetAuthService(),
		}
	})
	return user
}

// Register 验证用户名和密码长度以及邮箱格式， 验证邮箱验证码
// 验证用户名是否可用，通过从缓存读取email的邮箱验证码间接验证邮箱是否可用
func (u *UserController) Register(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "Register")
	if !pkg.GetConfig().Register {
		c.JSON(http.StatusBadRequest, global.ErrorOfForbidRegister)
		return
	}
	req := vo.UserModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	// 从缓存中使用邮箱作为key获取邮箱验证码与表单的邮箱验证码比对
	if !u.verify.VerifyEmailCaptcha(ctx, req.Email, req.Code) {
		c.JSON(http.StatusBadRequest, global.ErrorOfCaptcha)
		return
	}
	if len(req.Password) < global.UserMinLength || len(req.Password) > global.UserMaxLength {
		c.JSON(http.StatusBadRequest, global.ErrorOfPassword)
		return
	}
	if len(req.Username) < global.UserMinLength || len(req.Username) > global.UserMaxLength {
		c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
		return
	}
	if !utils.CheckEmailFormat(req.Email) {
		c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
		return
	}
	if _, res := u.user.GetUserByUsername(ctx, req.Username); res {
		c.JSON(http.StatusBadRequest, global.ErrorOfExistUsername)
		return
	}
	if err := u.user.Registry(ctx, req.Username, req.Password, req.Email); err != nil {
		log.WithError(err).Warnf("registry failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	u.verify.DeleteCaptcha(ctx, req.Email)
	c.Status(http.StatusOK)
}

// ResetPwd 邮箱链接下的重置密码
// 解析链接中的token，判断邮箱是否存在
// 验证密码长度，验证原密码和新密码是否一样
func (u *UserController) ResetPwd(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ResetPwd")
	req := vo.TokenModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	id, exp, err := encryption.ParseEmailToken(req.Token)
	if err != nil {
		log.WithError(err).Warnf("parse email token failed")
		c.JSON(http.StatusForbidden, global.ErrorOfExpectedLink)
		return
	} else if exp < utils.Epoch() {
		c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
		return
	}
	password := c.PostForm("password")
	if len(password) < global.UserMinLength || len(password) > global.UserMaxLength {
		c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
		return
	}
	if user, err := u.user.GetUserByID(ctx, id); err != nil {
		c.JSON(http.StatusBadRequest, global.ErrorOfExpectedLink)
		return
	} else if user.Password == password {
		c.JSON(http.StatusBadRequest, global.ErrorOfSamePassword)
		return
	}
	if err := u.user.ResetPwd(ctx, id, password); err != nil {
		log.WithError(err).Warnf("reset password failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.Status(http.StatusOK)
}

// ResetPwdByOrigin 用户信息界面的重置密码
// 验证原密码和新密码长度，验证原密码和新密码是否一样，验证原密码是否真实密码
func (u *UserController) ResetPwdByOrigin(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ResetPwdByOrigin")
	req := vo.ResetPasswordModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	user := u.GetAuth(c)
	if user.Password != encryption.EncryptPassword(req.Origin) {
		c.JSON(http.StatusBadRequest, global.ErrorOfWrongPassword)
		return
	} else if req.Origin == req.Reset {
		c.JSON(http.StatusBadRequest, global.ErrorOfSamePassword)
		return
	}
	if err := u.user.ResetPwd(ctx, user.ID, req.Reset); err != nil {
		log.WithError(err).Warnf("reset password failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.Status(http.StatusOK)
}

// ResetEmail 重置邮箱
// 验证邮箱格式以及和原邮箱是否一样，验证邮箱验证码
func (u *UserController) ResetEmail(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ResetPwdByOrigin")
	req := vo.ResetEmailModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	user := u.GetAuth(c)
	// 将email作为key从缓存中提取验证码比对
	if !u.verify.VerifyEmailCaptcha(ctx, req.Email, req.Code) {
		c.JSON(http.StatusBadRequest, global.ErrorOfCaptcha)
		return
	}
	if !utils.CheckEmailFormat(req.Email) {
		c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
		return
	} else if _, res := u.user.GetUserByEmail(ctx, req.Email); res {
		c.JSON(http.StatusBadRequest, global.ErrorOfExistEmail)
		return
	} else if err := u.user.ResetEmail(ctx, user.ID, req.Email); err != nil {
		log.WithError(err).Warnf("reset email failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	} else {
		u.verify.DeleteCaptcha(ctx, req.Email)
		user.Email = req.Email
		token, e := encryption.GenerateUserToken(user)
		if e != nil {
			log.WithError(err).Warnf("get token failed")
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
		if err = u.verify.SetToken(ctx, user.Name, token); err != nil {
			log.WithError(err).Warnf("set token failed")
		}
		c.JSON(http.StatusOK, map[string]interface{}{"token": token})
	}
}

// Unsubscribe 注销账户
func (u *UserController) Unsubscribe(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ResetPwdByOrigin")
	req := vo.UserModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	user := u.GetAuth(c)
	// 将email作为key从缓存中提取验证码比对
	if !u.verify.VerifyEmailCaptcha(ctx, user.Email, req.Code) {
		c.JSON(http.StatusBadRequest, global.ErrorOfCaptcha)
		return
	}
	// 从数据库中删除相关信息并从磁盘删除文件
	if err := u.user.UnsubscribeUser(ctx, user.ID, user.Storage.ID); err != nil {
		log.WithError(err).Warnf("unsubscribe failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.Status(http.StatusOK)
}
