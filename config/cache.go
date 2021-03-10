package config

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

var redisCache *redis.Client

// InitializeCache 用于Redis初始化
func InitializeCache() error {
	redisCache = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", GetConfig().Redis.Host, GetConfig().Redis.Port),
		Password:     GetConfig().Redis.Password,
		MinIdleConns: GetConfig().Redis.MinIdle,
		PoolSize:     GetConfig().Redis.Pool,
		IdleTimeout:  time.Duration(GetConfig().Redis.Timeout),
	})
	if err := redisCache.Ping().Err(); err != nil {
		return fmt.Errorf("redis 初始化失败: %v", err)
	}
	return nil
}

func GetCache() *redis.Client {
	return redisCache
}
