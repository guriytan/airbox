package pkg

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"

	"airbox/config"
)

var redisCache *redis.Client

// InitializeCache 用于Redis初始化
func InitializeCache() error {
	redisCache = redis.NewClient(&redis.Options{
		Addr:         config.GetConfig().Redis.Host,
		Password:     config.GetConfig().Redis.Password,
		MinIdleConns: 5,
		PoolSize:     20,
		IdleTimeout:  300 * time.Millisecond,
	})
	if err := redisCache.Ping().Err(); err != nil {
		return fmt.Errorf("redis 初始化失败: %v", err)
	}
	return nil
}

func GetCache() *redis.Client {
	return redisCache
}
