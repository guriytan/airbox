package controller

import (
	"airbox/config"
	"airbox/service"
	"airbox/utils"

	"net/http"

	"github.com/labstack/echo/v4"
)

// UserController is responsible for the request of user operation
type UserController struct {
	BaseController
	user   *service.UserService
	verify *service.AuthService
}

var user *UserController

// GetUserController return instance of UserController
func GetUserController() *UserController {
	if user == nil {
		user = &UserController{
			BaseController: getBaseController(),
			user:           service.GetUserService(),
			verify:         service.GetAuthService(),
		}
	}
	return user
}

// Register 验证用户名和密码长度以及邮箱格式， 验证邮箱验证码
// 验证用户名是否可用，通过从缓存读取email的邮箱验证码间接验证邮箱是否可用
func (u *UserController) Register(c echo.Context) error {
	if !config.Env.Register {
		return c.JSON(http.StatusBadRequest, "禁止注册")
	}
	email, code := c.FormValue("email"), c.FormValue("code")
	// 从缓存中使用邮箱作为key获取邮箱验证码与表单的邮箱验证码比对
	if !u.verify.VerifyEmailCaptcha(email, code) {
		return c.JSON(http.StatusBadRequest, "验证码错误")
	}
	password, username := c.FormValue("password"), c.FormValue("username")
	if len(password) < config.UserMinLength || len(password) > config.UserMaxLength {
		return c.JSON(http.StatusBadRequest, "密码应大于6位字符且少于18位字符")
	}
	if len(username) < config.UserMinLength || len(username) > config.UserMaxLength {
		return c.JSON(http.StatusBadRequest, "用户名应大于6位字符且少于18位字符")
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, "邮箱不能为空或格式错误")
	}
	if _, err := u.user.GetUserByUsername(username); err == nil {
		return c.JSON(http.StatusBadRequest, "用户已存在")
	}
	if err := u.user.Registry(username, password, email); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	u.verify.DeleteCaptcha(email)
	return c.NoContent(http.StatusOK)
}

// ResetPwd 邮箱链接下的重置密码
// 解析链接中的token，判断邮箱是否存在
// 验证密码长度，验证原密码和新密码是否一样
func (u *UserController) ResetPwd(c echo.Context) error {
	password := c.FormValue("password")
	if len(password) < config.UserMinLength || len(password) > config.UserMaxLength {
		return c.JSON(http.StatusBadRequest, "密码应大于6位字符且少于18位字符")
	}
	id := c.Get("id").(string)
	if user, err := u.user.GetUserByID(id); err != nil {
		return c.JSON(http.StatusBadRequest, "链接失效")
	} else if user.Password == utils.EncryptPassword(password) {
		return c.JSON(http.StatusBadRequest, "不能与原密码一致")
	}
	if err := u.user.ResetPwd(id, password); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

// ResetPwdByOrigin 用户信息界面的重置密码
// 验证原密码和新密码长度，验证原密码和新密码是否一样，验证原密码是否真实密码
func (u *UserController) ResetPwdByOrigin(c echo.Context) error {
	user := u.auth(c)
	origin, password := c.FormValue("origin"), c.FormValue("password")
	if user.Password != utils.EncryptPassword(origin) {
		return c.JSON(http.StatusBadRequest, "密码错误")
	} else if origin == password {
		return c.JSON(http.StatusBadRequest, "不能与原密码一致")
	}
	if err := u.user.ResetPwd(user.ID, password); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

// ResetEmail 重置邮箱
// 验证邮箱格式以及和原邮箱是否一样，验证邮箱验证码
func (u *UserController) ResetEmail(c echo.Context) error {
	user, email, code := u.auth(c), c.FormValue("email"), c.FormValue("code")
	// 将email作为key从缓存中提取验证码比对
	if !u.verify.VerifyEmailCaptcha(email, code) {
		return c.JSON(http.StatusBadRequest, "验证码错误")
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, "邮箱不能为空或格式错误")
	} else if _, err := u.user.GetUserByEmail(email); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusBadRequest, "邮箱已存在")
	} else if err := u.user.ResetEmail(user.ID, email); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		u.verify.DeleteCaptcha(email)
		user.Email = email
		token, e := utils.GenerateUserToken(user)
		if e != nil {
			c.Logger().Errorf("%s\n", e.Error())
			return c.JSON(http.StatusInternalServerError, e.Error())
		}
		if err = u.verify.SetToken(user.Name, token); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
		})
	}
}

// Unsubscribe 注销账户
func (u *UserController) Unsubscribe(c echo.Context) error {
	user := u.auth(c)
	// 将email作为key从缓存中提取验证码比对
	if code := c.QueryParam("code"); !u.verify.VerifyEmailCaptcha(user.Email, code) {
		return c.JSON(http.StatusBadRequest, config.ErrorOfCaptcha)
	}
	// 从数据库中删除相关信息并从磁盘删除文件
	if err := u.user.UnsubscribeUser(user.ID, user.Storage.ID); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
