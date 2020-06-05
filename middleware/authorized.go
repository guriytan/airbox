package middleware

import (
	"airbox/global"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"
	"net/http"

	"github.com/labstack/echo/v4"
)

var verify = service.GetAuthService()

// Login 拦截请求是否有权限
func Login(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			cookie, err := c.Cookie("air_box_token")
			if err != nil {
				return c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
			}
			token = cookie.Value
		}
		claims, exp, err := encryption.ParseUserToken(token)
		if err != nil {
			global.LOGGER.Printf("failed to parse token: %s\n", err.Error())
			return c.JSON(http.StatusForbidden, err.Error())
		}
		// 解析token获得claims对象后，取claims的username作为key从redis中获取token，若token不一致则认为该用户在其他设备登录
		// 因此需要重新登录
		if !verify.VerifyToken(claims.Name, token) {
			return c.JSON(http.StatusUnauthorized, global.ErrorSSO)
		}
		// token过期
		if exp < utils.Epoch() {
			return c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
		}
		c.Set("Authorization", claims)
		return next(c)
	}
}

// CheckLink 拦截重置密码的链接是否有效
func CheckLink(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		id, exp, err := encryption.ParseEmailToken(token)
		if err != nil {
			global.LOGGER.Printf("failed to parse token: %s\n", err.Error())
			return c.JSON(http.StatusForbidden, err.Error())
		} else if exp < utils.Epoch() {
			return c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
		}
		c.Set("id", id)
		return next(c)
	}
}
