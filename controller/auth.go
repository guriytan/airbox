package controller

import (
	"net/http"

	"airbox/config"
	"airbox/global"
	"airbox/logger"
	"airbox/model"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	BaseController
	file   *service.FileService
	user   *service.UserService
	verify *service.AuthService
}

var auth *AuthController

func GetAuthController() *AuthController {
	if auth == nil {
		auth = &AuthController{
			BaseController: getBaseController(),
			file:           service.GetFileService(),
			user:           service.GetUserService(),
			verify:         service.GetAuthService(),
		}
	}
	return auth
}

// LoginToken 可通过输入用户名或者邮箱进行登录
// 需要验证邮箱格式，用户名和密码长度以及验证码
func (auth *AuthController) LoginToken(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "LoginToken")
	user, password := c.FormValue("user"), c.FormValue("password")
	if !utils.CheckEmailFormat(user) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(user) < global.UserMinLength || len(user) > global.UserMaxLength {
			return c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
		}
	}
	if len(password) < global.UserMinLength || len(password) > global.UserMaxLength {
		return c.JSON(http.StatusBadRequest, global.ErrorOfPassword)
	}
	if user, err := auth.user.Login(ctx, user, password); err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	} else {
		token, e := encryption.GenerateUserToken(user)
		if e != nil {
			log.Infof("%+v\n", err)
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		if err = auth.verify.SetToken(ctx, user.Name, token); err != nil {
			log.Infof("%+v\n", err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
		})
	}
}

// UnsubscribeCode 申请注销账户，向用户邮箱发送验证码邮件验证权限
func (auth *AuthController) UnsubscribeCode(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "UnsubscribeCode")
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(ctx, auth.auth(c).Email); err != nil {
			log.Infof("%+v\n", err)
		}
	}()
	return c.NoContent(http.StatusOK)
}

// ShareLink 获取分享文件的link
func (auth *AuthController) ShareLink(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "ShareLink")
	id := c.FormValue("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, global.ErrorOfFileID)
	}
	fileByID, err := auth.file.GetFileByID(ctx, id)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	token, err := encryption.GenerateShareToken(fileByID.ID)
	if err != nil {
		log.Infof("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"link": token,
	})
}

// RegisterCode 发送注册需要的邮箱验证码
// 验证用户名长度和邮箱格式， 验证用户名和邮箱是否唯一
func (auth *AuthController) RegisterCode(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "RegisterCode")
	if !config.GetConfig().Register {
		return c.JSON(http.StatusBadRequest, global.ErrorOfForbidRegister)
	}
	username, email := c.FormValue("username"), c.FormValue("email")
	if len(username) < global.UserMinLength || len(username) > global.UserMaxLength {
		return c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
	}
	if _, res := auth.user.GetUserByUsername(ctx, username); !res {
		return c.JSON(http.StatusBadRequest, global.ErrorOfExistUsername)
	}
	if _, res := auth.user.GetUserByEmail(ctx, email); !res {
		return c.JSON(http.StatusBadRequest, global.ErrorOfExistEmail)
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(ctx, email); err != nil {
			log.Infof("%+v\n", err)
		}
	}()
	return c.NoContent(http.StatusOK)
}

// PasswordCode 忘记密码，发送重置密码的链接至邮箱，可输入用户名或邮箱
// 验证邮箱格式，若不是邮箱格式则为用户名，然后验证用户名长度
// 若为用户名则访问数据库获得邮箱
// 发送重置密码链接至邮箱
func (auth *AuthController) PasswordCode(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "PasswordCode")
	info := c.FormValue("info")
	var user *model.User
	var res bool
	if !utils.CheckEmailFormat(info) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(info) < global.UserMinLength || len(info) > global.UserMaxLength {
			return c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
		}
		// 从数据库获取用户的邮箱
		user, res = auth.user.GetUserByUsername(ctx, info)
		if !res {
			return c.NoContent(http.StatusOK)
		}
		info = user.Email
	} else if user, res = auth.user.GetUserByEmail(ctx, info); !res {
		// 当输入为邮箱时，检查邮箱是否存在
		return c.NoContent(http.StatusOK)
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendResetLink(user.ID, info); err != nil {
			log.Infof("%+v\n", err)
		}
	}()
	return c.NoContent(http.StatusOK)
}

// EmailCode 发送重置邮箱需要的邮箱验证码
// 直接发送邮箱验证码
func (auth *AuthController) EmailCode(c echo.Context) error {
	ctx := c.Request().Context()

	log := logger.GetLogger(ctx, "EmailCode")
	user, email, password := auth.auth(c), c.FormValue("email"), c.FormValue("password")
	if user.Password != encryption.EncryptPassword(password) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfWrongPassword)
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
	} else if user.Email == email {
		return c.JSON(http.StatusBadRequest, global.ErrorOfSameEmail)
	} else if _, res := auth.user.GetUserByEmail(ctx, email); !res {
		return c.JSON(http.StatusBadRequest, global.ErrorOfExistEmail)
	} else {
		// 发送验证码至邮箱
		go func() {
			if err := auth.verify.SendCaptcha(ctx, email); err != nil {
				log.Infof("%+v\n", err)
			}
		}()
	}
	return c.NoContent(http.StatusOK)
}
