package middleware

import (
	"net/http"

	"airbox/global"
	"airbox/logger"
	"airbox/model/do"
	"airbox/service"
	"airbox/utils"
	"airbox/utils/encryption"

	"github.com/gin-gonic/gin"
)

// Login 拦截请求是否有权限
func Login(c *gin.Context) {
	log := logger.GetLogger(c, "Login")
	token := c.GetHeader("Authorization")
	if len(token) == 0 {
		cookie, err := c.Cookie("air_box_token")
		if err != nil {
			c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
			return
		}
		token = cookie
	}
	var claims do.User
	exp, err := encryption.ParseUserToken(token, &claims)
	if err != nil {
		log.Infof("failed to parse token: %+v\n", err)
		c.JSON(http.StatusForbidden, global.ErrorWithoutToken)
		return
	}
	// 解析token获得claims对象后，取claims的username作为key从redis中获取token，若token不一致则认为该用户在其他设备登录
	// 因此需要重新登录
	if !service.GetAuthService().VerifyToken(c, claims.Name, token) {
		c.JSON(http.StatusUnauthorized, global.ErrorSSO)
		return
	}
	// token过期
	if exp < utils.Epoch() {
		c.JSON(http.StatusUnauthorized, global.ErrorOutOfDated)
		return
	}
	c.Set(global.KeyAuthorization, &claims)
	c.Set(global.KeyUserID, claims.ID)

	c.Next()
}
