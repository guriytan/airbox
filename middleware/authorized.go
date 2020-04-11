package middleware

import (
	"airbox/config"
	"airbox/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ProlongToken 用于延长用户Token
//func ProlongToken(w http.ResponseWriter, claims *Claims) {
//	if claims.ExpiresAt-time.Now().Unix() < int64(ProlongDuration.Seconds()) {
//		token, err := GenerateToken(&claims.User)
//		if err == nil {
//			http.SetCookie(w, &http.Cookie{
//				Name:   "session-id",
//				Value:  token,
//				MaxAge: 1800,
//			})
//		} else {
//			log.Printf("prolong the expiry date of token error: %s", err.Error())
//		}
//	}
//}

// Login 拦截请求是否有权限
func Login(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == "OPTIONS" {
			return next(c)
		}
		token, err := c.Cookie("air_box_token")
		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    config.CodeErrorOfAuthority,
				"warning": err.Error(),
			})
		}
		claims, exp, err := utils.ParseUserToken(token.Value)
		if err != nil {
			c.Logger().Warnf("failed to parse token: a", err)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    config.CodeErrorOfAuthority,
				"warning": err.Error(),
			})
		}
		if exp < utils.Epoch() {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    config.CodeErrorOfAuthority,
				"warning": config.ErrorOutOfDated,
			})
		}
		c.Set("authority", claims)
		return next(c)
	}
}

// CheckLink 拦截重置密码的链接是否有效
func CheckLink(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		id, exp, err := utils.ParseEmailToken(token)
		if err != nil {
			c.Logger().Warnf("failed to parse token: a", err)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    config.CodeErrorOfServer,
				"warning": err.Error(),
			})
		} else if exp < utils.Epoch() {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    config.CodeErrorOfAuthority,
				"warning": config.ErrorOutOfDated,
			})
		}
		c.Set("id", id)
		return next(c)
	}
}
