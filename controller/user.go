package controller

import (
	"airbox/global"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"

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
	if !global.Env.Register {
		return c.JSON(http.StatusBadRequest, global.ErrorOfForbidRegister)
	}
	email, code := c.FormValue("email"), c.FormValue("code")
	// 从缓存中使用邮箱作为key获取邮箱验证码与表单的邮箱验证码比对
	if !u.verify.VerifyEmailCaptcha(email, code) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfCaptcha)
	}
	password, username := c.FormValue("password"), c.FormValue("username")
	if len(password) < global.UserMinLength || len(password) > global.UserMaxLength {
		return c.JSON(http.StatusBadRequest, global.ErrorOfPassword)
	}
	if len(username) < global.UserMinLength || len(username) > global.UserMaxLength {
		return c.JSON(http.StatusBadRequest, global.ErrorOfUsername)
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
	}
	if _, res := u.user.GetUserByUsername(username); !res {
		return c.JSON(http.StatusBadRequest, global.ErrorOfExistUsername)
	}
	if err := u.user.Registry(username, password, email); err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	u.verify.DeleteCaptcha(email)
	return c.NoContent(http.StatusOK)
}

// ResetPwd 邮箱链接下的重置密码
// 解析链接中的token，判断邮箱是否存在
// 验证密码长度，验证原密码和新密码是否一样
func (u *UserController) ResetPwd(c echo.Context) error {
	password := c.FormValue("password")
	if len(password) < global.UserMinLength || len(password) > global.UserMaxLength {
		return c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
	}
	id := c.Get("id").(string)
	if user, err := u.user.GetUserByID(id); err != nil {
		return c.JSON(http.StatusBadRequest, global.ErrorOfExpectedLink)
	} else if user.Password == encryption.EncryptPassword(password) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfSamePassword)
	}
	if err := u.user.ResetPwd(id, password); err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.NoContent(http.StatusOK)
}

// ResetPwdByOrigin 用户信息界面的重置密码
// 验证原密码和新密码长度，验证原密码和新密码是否一样，验证原密码是否真实密码
func (u *UserController) ResetPwdByOrigin(c echo.Context) error {
	user := u.auth(c)
	origin, password := c.FormValue("origin"), c.FormValue("password")
	if user.Password != encryption.EncryptPassword(origin) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfWrongPassword)
	} else if origin == password {
		return c.JSON(http.StatusBadRequest, global.ErrorOfSamePassword)
	}
	if err := u.user.ResetPwd(user.ID, password); err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.NoContent(http.StatusOK)
}

// ResetEmail 重置邮箱
// 验证邮箱格式以及和原邮箱是否一样，验证邮箱验证码
func (u *UserController) ResetEmail(c echo.Context) error {
	user, email, code := u.auth(c), c.FormValue("email"), c.FormValue("code")
	// 将email作为key从缓存中提取验证码比对
	if !u.verify.VerifyEmailCaptcha(email, code) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfCaptcha)
	}
	if !utils.CheckEmailFormat(email) {
		return c.JSON(http.StatusBadRequest, global.ErrorOfEmail)
	} else if _, res := u.user.GetUserByEmail(email); !res {
		return c.JSON(http.StatusBadRequest, global.ErrorOfExistEmail)
	} else if err := u.user.ResetEmail(user.ID, email); err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	} else {
		u.verify.DeleteCaptcha(email)
		user.Email = email
		token, e := encryption.GenerateUserToken(user)
		if e != nil {
			global.LOGGER.Printf("%s\n", e.Error())
			return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		}
		if err = u.verify.SetToken(user.Name, token); err != nil {
			global.LOGGER.Printf("%s\n", err.Error())
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
		return c.JSON(http.StatusBadRequest, global.ErrorOfCaptcha)
	}
	// 从数据库中删除相关信息并从磁盘删除文件
	if err := u.user.UnsubscribeUser(user.ID, user.Storage.ID); err != nil {
		global.LOGGER.Printf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.NoContent(http.StatusOK)
}
