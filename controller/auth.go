package controller

import (
	"net/http"
	"sync"

	"airbox/config"
	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/model/vo"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	BaseController
	file   *service.FileService
	user   *service.UserService
	verify *service.AuthService
}

var (
	auth     *AuthController
	authOnce sync.Once
)

func GetAuthController() *AuthController {
	authOnce.Do(func() {
		auth = &AuthController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			user:           service.GetUserService(),
			verify:         service.GetAuthService(),
		}
	})
	return auth
}

// LoginToken 可通过输入用户名或者邮箱进行登录
// 需要验证邮箱格式，用户名和密码长度以及验证码
func (auth *AuthController) LoginToken(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "LoginToken")
	req := vo.LoginModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if !utils.CheckEmailFormat(req.UserKey) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(req.UserKey) < global.UserMinLength || len(req.UserKey) > global.UserMaxLength {
			c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
			return
		}
	}
	if len(req.Password) < global.UserMinLength || len(req.Password) > global.UserMaxLength {
		c.JSON(http.StatusBadRequest, global.ErrorOfPassword)
		return
	}
	if user, err := auth.user.Login(ctx, req.UserKey, req.Password); err != nil {
		log.WithError(err).Warnf("login failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	} else {
		token, e := encryption.GenerateUserToken(user)
		if e != nil {
			log.WithError(err).Warnf("get token failed")
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
		if err = auth.verify.SetToken(ctx, user.Name, token); err != nil {
			log.WithError(err).Warnf("set token failed")
		}
		c.JSON(http.StatusOK, map[string]interface{}{"token": token})
	}
}

// UnsubscribeCode 申请注销账户，向用户邮箱发送验证码邮件验证权限
func (auth *AuthController) UnsubscribeCode(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "UnsubscribeCode")
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(ctx, auth.GetAuth(c).Email); err != nil {
			log.WithError(err).Warnf("send captcha failed")
		}
	}()
	c.Status(http.StatusOK)
}

// ShareLink 获取分享文件的link
func (auth *AuthController) ShareLink(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "ShareLink")
	req := vo.ShareModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if len(req.FileID) == 0 {
		c.JSON(http.StatusBadRequest, global.ErrorOfFileID)
		return
	}
	fileByID, err := auth.file.GetFileByID(ctx, req.FileID)
	if err != nil {
		log.WithError(err).Warnf("get file: %v failed", req.FileID)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	token, err := encryption.GenerateShareToken(fileByID.ID)
	if err != nil {
		log.WithError(err).Warnf("get file token failed")
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"link": token})
}

// RegisterCode 发送注册需要的邮箱验证码
// 验证用户名长度和邮箱格式， 验证用户名和邮箱是否唯一
func (auth *AuthController) RegisterCode(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "RegisterCode")
	if !pkg.GetConfig().Register {
		c.JSON(http.StatusBadRequest, global.ErrorOfForbidRegister)
		return
	}
	req := vo.UserModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
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
	if _, res := auth.user.GetUserByUsername(ctx, req.Username); res {
		c.JSON(http.StatusBadRequest, global.ErrorOfExistUsername)
		return
	}
	if _, res := auth.user.GetUserByEmail(ctx, req.Email); res {
		c.JSON(http.StatusBadRequest, global.ErrorOfExistEmail)
		return
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(ctx, req.Email); err != nil {
			log.WithError(err).Warnf("send captcha failed")
		}
	}()
	c.Status(http.StatusOK)
}

// PasswordCode 忘记密码，发送重置密码的链接至邮箱，可输入用户名或邮箱
// 验证邮箱格式，若不是邮箱格式则为用户名，然后验证用户名长度
// 若为用户名则访问数据库获得邮箱
// 发送重置密码链接至邮箱
func (auth *AuthController) PasswordCode(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(ctx, "PasswordCode")
	req := vo.LoginModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var user *do.User
	var res bool
	if !utils.CheckEmailFormat(req.UserKey) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(req.UserKey) < global.UserMinLength || len(req.UserKey) > global.UserMaxLength {
			c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
			return
		}
		// 从数据库获取用户的邮箱
		user, res = auth.user.GetUserByUsername(ctx, req.UserKey)
		if res {
			c.Status(http.StatusOK)
			return
		}
		req.UserKey = user.Email
	} else if user, res = auth.user.GetUserByEmail(ctx, req.UserKey); res {
		// 当输入为邮箱时，检查邮箱是否存在
		c.Status(http.StatusOK)
		return
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendResetLink(ctx, user.ID, req.UserKey); err != nil {
			log.WithError(err).Warnf("send reset link failed")
		}
	}()
	c.Status(http.StatusOK)
}

// EmailCode 发送重置邮箱需要的邮箱验证码
// 直接发送邮箱验证码
func (auth *AuthController) EmailCode(c *gin.Context) {
	ctx := utils.CopyCtx(c)

	log := logger.GetLogger(c, "EmailCode")
	req := vo.UserModel{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	user := auth.GetAuth(c)
	if user.Password != req.Password {
		c.JSON(http.StatusBadRequest, global.ErrorOfWrongPassword)
		return
	}
	if !utils.CheckEmailFormat(req.Email) {
		c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
		return
	}
	if user.Email == req.Email {
		c.JSON(http.StatusBadRequest, global.ErrorOfSameEmail)
		return
	}
	if _, res := auth.user.GetUserByEmail(ctx, req.Email); res {
		c.JSON(http.StatusBadRequest, global.ErrorOfExistEmail)
		return
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(ctx, req.Email); err != nil {
			log.WithError(err).Warnf("send captcha failed")
		}
	}()
	c.Status(http.StatusOK)
}
