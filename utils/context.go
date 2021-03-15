package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"airbox/global"
)

func IsDev(ctx context.Context) bool {
	if dev, ok := ctx.Value(global.KeyDev).(string); ok && dev == "true" {
		return true
	}
	return false
}

func CopyCtx(c *gin.Context) context.Context {
	ctx := &commonCtx{values: map[string]interface{}{}}
	for key, value := range c.Keys {
		ctx.values[key] = value
	}
	for key, strings := range c.Request.Header {
		if len(strings) > 0 {
			ctx.values[key] = strings[0]
		}
	}
	return ctx
}

type commonCtx struct {
	values map[string]interface{}
}

func (c *commonCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *commonCtx) Done() <-chan struct{} {
	return nil
}

func (c *commonCtx) Err() error {
	return nil
}

func (c *commonCtx) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		val, _ := c.values[keyAsString]
		return val
	}
	return nil
}
