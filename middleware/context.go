package middleware

import (
	"context"

	"airbox/global"
	"airbox/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// InjectContext 注入常见上下文
func InjectContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		requestID := c.Request().Header.Get(global.KeyRequestID)
		if len(requestID) == 0 {
			requestID = uuid.New().String()
			c.Request().Header.Set(global.KeyRequestID, requestID)
		}

		ctx = context.WithValue(ctx, global.KeyRequestID, requestID)
		ctx = context.WithValue(ctx, global.KeyIP, c.RealIP())

		user, ok := c.Get("Authorization").(*model.User)
		if ok {
			ctx = context.WithValue(ctx, global.KeyUserID, user.ID)
		}

		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
