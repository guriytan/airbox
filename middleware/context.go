package middleware

import (
	"airbox/global"
	"airbox/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InjectContext 注入常见上下文
func InjectContext(c *gin.Context) {

	requestID := c.GetHeader(global.KeyRequestID)
	if len(requestID) == 0 {
		requestID = uuid.New().String()
		c.Header(global.KeyRequestID, requestID)
	}

	c.Set(global.KeyRequestID, requestID)
	c.Set(global.KeyIP, c.ClientIP())

	user, ok := c.Get("Authorization")
	if ok {
		c.Set(global.KeyUserID, user.(*model.User).ID)
	}

	c.Next()
}
