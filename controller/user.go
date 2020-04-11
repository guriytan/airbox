package controller

import (
	"airbox/config"
	"airbox/model"
	"airbox/service"
	"airbox/utils"

	"net/http"

	"github.com/labstack/echo/v4"
)

// UserController is responsible for the request of user operation
type UserController struct {
	*BaseController
	user *service.UserService
}

var user *UserController

// GetUserService return instance of UserController
func GetUserService() *UserController {
	if user == nil {
		user = &UserController{
			BaseController: GetBaseController(),
			user:           service.GetUserService(),
		}
	}
	return user
}

// Login 可通过输入用户名或者邮箱进行登录
// 需要验证邮箱格式，用户名和密码长度以及验证码
func (u *UserController) Login(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	user, password := c.FormValue("user"), c.FormValue("password")
	username, email := "", ""
	if !utils.CheckEmailFormat(user) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(user) < config.UserMinLength || len(user) > config.UserMaxLength {
			data["warning"] = "用户名长度应在6位至18位之间"
			return c.JSON(http.StatusOK, data)
		} else {
			username = user
		}
	} else {
		email = user
	}
	if len(password) < config.UserMinLength || len(password) > config.UserMaxLength {
		data["warning"] = "密码长度应在6位至18位之间"
		return c.JSON(http.StatusOK, data)
	}
	if user, err := u.user.Login(username, email, password); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		data["warning"] = err.Error()
		data["code"] = config.CodeErrorOfServer
		return c.JSON(http.StatusOK, data)
	} else {
		token, e := utils.GenerateUserToken(user)
		if e != nil {
			c.Logger().Errorf("%s\n", e.Error())
			data["warning"] = e.Error()
			data["code"] = config.CodeErrorOfServer
			return c.JSON(http.StatusOK, data)
		}
		data["code"] = config.CodeSuccess
		data["token"] = token
		return c.JSON(http.StatusOK, data)
	}
}

// Register 验证用户名和密码长度以及邮箱格式， 验证邮箱验证码
// 验证用户名是否可用，通过从缓存读取email的邮箱验证码间接验证邮箱是否可用
func (u *UserController) Register(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	if !config.Env.Register {
		data["warning"] = "禁止注册"
		return c.JSON(http.StatusOK, data)
	}
	email, code := c.FormValue("email"), c.FormValue("code")
	// 从缓存中使用邮箱作为key获取邮箱验证码与表单的邮箱验证码比对
	if !u.user.VerifyEmailCaptcha(email, code) {
		data["warning"] = "验证码错误"
		return c.JSON(http.StatusOK, data)
	}
	password, username := c.FormValue("password"), c.FormValue("username")
	if len(password) < config.UserMinLength || len(password) > config.UserMaxLength {
		data["warning"] = "密码应大于6位字符且少于18位字符"
		return c.JSON(http.StatusOK, data)
	}
	if len(username) < config.UserMinLength || len(username) > config.UserMaxLength {
		data["warning"] = "用户名应大于6位字符且少于18位字符"
		return c.JSON(http.StatusOK, data)
	}
	if !utils.CheckEmailFormat(email) {
		data["warning"] = "邮箱不能为空或格式错误"
		return c.JSON(http.StatusOK, data)
	}
	if _, err := u.user.GetUserByUsername(username); err == nil {
		data["warning"] = "用户已存在"
		return c.JSON(http.StatusOK, data)
	}
	if err := u.user.Registry(username, password, email); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		data["warning"] = err.Error()
		data["code"] = config.CodeErrorOfServer
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// SendRegisterCode 发送注册需要的邮箱验证码
// 验证用户名长度和邮箱格式， 验证用户名和邮箱是否唯一
func (u *UserController) SendRegisterCode(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	username, email := c.FormValue("username"), c.FormValue("email")
	if len(username) < config.UserMinLength || len(username) > config.UserMaxLength {
		data["warning"] = "用户名应大于6位字符且少于18位字符"
		return c.JSON(http.StatusOK, data)
	}
	if !utils.CheckEmailFormat(email) {
		data["warning"] = "邮箱不能为空或格式错误"
		return c.JSON(http.StatusOK, data)
	}
	if _, err := u.user.GetUserByUsername(username); err == nil {
		data["warning"] = "用户已存在"
		return c.JSON(http.StatusOK, data)
	}
	if _, err := u.user.GetUserByEmail(email); err == nil {
		data["warning"] = "邮箱已存在"
		return c.JSON(http.StatusOK, data)
	}
	// 发送验证码至邮箱
	go func() {
		if err := u.user.SendCaptcha(email); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
	}()
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// ForgetPwd 忘记密码，发送重置密码的链接至邮箱，可输入用户名或邮箱
// 验证邮箱格式，若不是邮箱格式则为用户名，然后验证用户名长度
// 若为用户名则访问数据库获得邮箱
// 发送重置密码链接至邮箱
func (u *UserController) ForgetPwd(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeSuccess,
	}
	info := c.FormValue("info")
	var user *model.User
	var err error
	if !utils.CheckEmailFormat(info) {
		// 用户输入用户名进行登录，判断用户名长度
		if len(info) < config.UserMinLength || len(info) > config.UserMaxLength {
			data["code"] = config.CodeErrorOfRequest
			data["warning"] = "用户名长度应在6位至18位之间"
			return c.JSON(http.StatusOK, data)
		}
		// 从数据库获取用户的邮箱
		user, err = u.user.GetUserByUsername(info)
		if err != nil {
			return c.JSON(http.StatusOK, data)
		}
		info = user.Email
	} else if user, err = u.user.GetUserByEmail(info); err != nil {
		// 当输入为邮箱时，检查邮箱是否存在
		return c.JSON(http.StatusOK, data)
	}
	// 发送验证码至邮箱
	go func() {
		if err := u.user.SendResetLink(user.ID, info); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
		}
	}()
	return c.JSON(http.StatusOK, data)
}

// ResetPwd 邮箱链接下的重置密码
// 解析链接中的token，判断邮箱是否存在
// 验证密码长度，验证原密码和新密码是否一样
func (u *UserController) ResetPwd(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	password := c.FormValue("password")
	if len(password) < config.UserMinLength || len(password) > config.UserMaxLength {
		data["warning"] = "密码应大于6位字符且少于18位字符"
		return c.JSON(http.StatusOK, data)
	}
	id := c.Get("id").(string)
	if user, err := u.user.GetUserByID(id); err != nil {
		data["warning"] = "链接失效"
		return c.JSON(http.StatusOK, data)
	} else if user.Password == utils.EncryptPassword(password) {
		data["warning"] = "不能与原密码一致"
		return c.JSON(http.StatusOK, data)
	}
	if err := u.user.ResetPwd(id, password); err != nil {
		data["warning"] = "重置密码错误"
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// ResetPwdByOrigin 用户信息界面的重置密码
// 验证原密码和新密码长度，验证原密码和新密码是否一样，验证原密码是否真实密码
func (u *UserController) ResetPwdByOrigin(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	user := u.Auth(c)
	origin, password := c.FormValue("origin"), c.FormValue("password")
	if user.Password != utils.EncryptPassword(origin) {
		data["warning"] = "密码错误"
		return c.JSON(http.StatusOK, data)
	} else if origin == password {
		data["warning"] = "不能与原密码一致"
		return c.JSON(http.StatusOK, data)
	}
	if err := u.user.ResetPwd(user.ID, password); err != nil {
		data["warning"] = "重置密码错误"
		return c.JSON(http.StatusOK, data)
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}

// ResetEmail 重置邮箱
// 验证邮箱格式以及和原邮箱是否一样，验证邮箱验证码
func (u *UserController) ResetEmail(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	user := u.Auth(c)
	email, code := c.FormValue("email"), c.FormValue("code")
	// 将email作为key从缓存中提取验证码比对
	if !u.user.VerifyEmailCaptcha(email, code) {
		data["warning"] = "验证码错误"
		return c.JSON(http.StatusOK, data)
	}
	if !utils.CheckEmailFormat(email) {
		data["warning"] = "邮箱不能为空或格式错误"
		return c.JSON(http.StatusOK, data)
	} else if _, err := u.user.GetUserByEmail(email); err != nil {
		data["warning"] = "邮箱已存在"
		return c.JSON(http.StatusOK, data)
	} else if err := u.user.ResetEmail(user.ID, email); err != nil {
		data["warning"] = "更换邮箱失败"
		return c.JSON(http.StatusOK, data)
	} else {
		user.Email = email
		token, e := utils.GenerateUserToken(user)
		if e != nil {
			c.Logger().Errorf("%s\n", e.Error())
			data["warning"] = e.Error()
			data["code"] = config.CodeErrorOfServer
			return c.JSON(http.StatusOK, data)
		}
		data["code"] = config.CodeSuccess
		data["token"] = token
		return c.JSON(http.StatusOK, data)
	}
}

// SendResetEmailCode 发送重置邮箱需要的邮箱验证码
// 直接发送邮箱验证码
func (u *UserController) SendResetEmailCode(c echo.Context) error {
	data := map[string]interface{}{
		"code": config.CodeErrorOfRequest,
	}
	user := u.Auth(c)
	email, password := c.FormValue("email"), c.FormValue("password")
	if user.Password != utils.EncryptPassword(password) {
		data["warning"] = "密码错误"
		return c.JSON(http.StatusOK, data)
	}
	if !utils.CheckEmailFormat(email) {
		data["warning"] = "邮箱不能为空或格式错误"
		return c.JSON(http.StatusOK, data)
	} else if user.Email == email {
		data["warning"] = "新邮箱不能与原邮箱一致"
		return c.JSON(http.StatusOK, data)
	} else if _, err := u.user.GetUserByEmail(email); err != nil {
		data["warning"] = "邮箱已存在"
		return c.JSON(http.StatusOK, data)
	} else {
		// 发送验证码至邮箱
		go func() {
			if err := u.user.SendCaptcha(email); err != nil {
				c.Logger().Errorf("%s\n", err.Error())
			}
		}()
	}
	data["code"] = config.CodeSuccess
	return c.JSON(http.StatusOK, data)
}
