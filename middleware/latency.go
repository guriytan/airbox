package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"airbox/logger"
)

// Latency 请求入口日志和耗时
func Latency(c *gin.Context) {
	start := time.Now()
	defer func() {
		log := logger.GetLogger(c, "Latency")
		log.Infof("request path: %s, handle cost: %d ms", c.FullPath(), time.Since(start).Milliseconds())
	}()
	c.Next()
}
