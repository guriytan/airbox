package middleware

import (
	"airbox/config"
	"airbox/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Login 拦截请求是否有权限
func Login(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			cookie, err := c.Cookie("air_box_token")
			if err != nil {
				return c.JSON(http.StatusForbidden, config.ErrorWithoutToken)
			}
			token = cookie.Value
		}
		claims, exp, err := utils.ParseUserToken(token)
		if err != nil {
			c.Logger().Warnf("failed to parse token: %s\n", err.Error())
			return c.JSON(http.StatusForbidden, err.Error())
		}
		if exp < utils.Epoch() {
			return c.JSON(http.StatusUnauthorized, config.ErrorOutOfDated)
		}
		c.Set("Authorization", claims)
		return next(c)
	}
}

// CheckLink 拦截重置密码的链接是否有效
func CheckLink(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		id, exp, err := utils.ParseEmailToken(token)
		if err != nil {
			c.Logger().Warnf("failed to parse token: %s\n", err.Error())
			return c.JSON(http.StatusForbidden, err.Error())
		} else if exp < utils.Epoch() {
			return c.JSON(http.StatusUnauthorized, config.ErrorOutOfDated)
		}
		c.Set("id", id)
		return next(c)
	}
}
