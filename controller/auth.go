package controller

import (
	"airbox/config"
	"airbox/model"
	"airbox/service"
	"airbox/utils"
	"github.com/labstack/echo/v4"
	"net/http"
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
	user, password := c.FormValue("user"), c.FormValue("password")
	if !utils.CheckEmailFormat(user) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(user) < config.UserMinLength || len(user) > config.UserMaxLength {
			return c.JSON(http.StatusBadRequest, "用户名长度应在6位至18位之间")
		}
	}
	if len(password) < config.UserMinLength || len(password) > config.UserMaxLength {
		return c.JSON(http.StatusBadRequest, "密码长度应在6位至18位之间")
	}
	if user, err := auth.user.Login(user, password); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		token, e := utils.GenerateUserToken(user)
		if e != nil {
			c.Logger().Errorf("%s\n", e.Error())
			return c.JSON(http.StatusInternalServerError, e.Error())
		}
		if err = auth.verify.SetToken(user.Name, token); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
		})
	}
}

// UnsubscribeCode 申请注销账户，向用户邮箱发送验证码邮件验证权限
func (auth *AuthController) UnsubscribeCode(c echo.Context) error {
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(auth.auth(c).Email); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
	}()
	return c.NoContent(http.StatusOK)
}

// ShareLink 获取分享文件的link
func (auth *AuthController) ShareLink(c echo.Context) error {
	id := c.FormValue("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, "缺少文件id")
	}
	fileByID, err := auth.file.GetFileByID(id)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	token, err := utils.GenerateShareToken(fileByID.ID)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"link": token,
	})
}

// RegisterCode 发送注册需要的邮箱验证码
// 验证用户名长度和邮箱格式， 验证用户名和邮箱是否唯一
func (auth *AuthController) RegisterCode(c echo.Context) error {
	if !config.Env.Register {
		return c.JSON(http.StatusBadRequest, "禁止注册")
	}
	username, email := c.FormValue("username"), c.FormValue("email")
	if len(username) < config.UserMinLength || len(username) > config.UserMaxLength {
		return c.JSON(http.StatusBadRequest, "用户名应大于6位字符且少于18位字符")
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, "邮箱不能为空或格式错误")
	}
	if _, err := auth.user.GetUserByUsername(username); err == nil {
		return c.JSON(http.StatusBadRequest, "用户已存在")
	}
	if _, err := auth.user.GetUserByEmail(email); err == nil {
		return c.JSON(http.StatusBadRequest, "邮箱已存在")
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendCaptcha(email); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
	}()
	return c.NoContent(http.StatusOK)
}

// PasswordCode 忘记密码，发送重置密码的链接至邮箱，可输入用户名或邮箱
// 验证邮箱格式，若不是邮箱格式则为用户名，然后验证用户名长度
// 若为用户名则访问数据库获得邮箱
// 发送重置密码链接至邮箱
func (auth *AuthController) PasswordCode(c echo.Context) error {
	info := c.FormValue("info")
	var user *model.User
	var err error
	if !utils.CheckEmailFormat(info) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(info) < config.UserMinLength || len(info) > config.UserMaxLength {
			return c.JSON(http.StatusBadRequest, "用户名长度应在6位至18位之间")
		}
		// 从数据库获取用户的邮箱
		user, err = auth.user.GetUserByUsername(info)
		if err != nil {
			c.Logger().Errorf("%s\n", err.Error())
			return c.NoContent(http.StatusOK)
		}
		info = user.Email
	} else if user, err = auth.user.GetUserByEmail(info); err != nil {
		// 当输入为邮箱时，检查邮箱是否存在
		c.Logger().Errorf("%s\n", err.Error())
		return c.NoContent(http.StatusOK)
	}
	// 发送验证码至邮箱
	go func() {
		if err := auth.verify.SendResetLink(user.ID, info); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
	}()
	return c.NoContent(http.StatusOK)
}

// EmailCode 发送重置邮箱需要的邮箱验证码
// 直接发送邮箱验证码
func (auth *AuthController) EmailCode(c echo.Context) error {
	user, email, password := auth.auth(c), c.FormValue("email"), c.FormValue("password")
	if user.Password != utils.EncryptPassword(password) {
		return c.JSON(http.StatusBadRequest, "密码错误")
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, "邮箱不能为空或格式错误")
	} else if user.Email == email {
		return c.JSON(http.StatusBadRequest, "新邮箱不能与原邮箱一致")
	} else if _, err := auth.user.GetUserByEmail(email); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusBadRequest, "邮箱已存在")
	} else {
		// 发送验证码至邮箱
		go func() {
			if err := auth.verify.SendCaptcha(email); err != nil {
				c.Logger().Errorf("%s\n", err.Error())
			}
		}()
	}
	return c.NoContent(http.StatusOK)
}
