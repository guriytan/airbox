package config

import (
	"airbox/global"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

var redisCache *redis.Client

// InitializeRedis 用于Redis初始化
func InitializeRedis() {
	redisCache = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", global.Env.Redis.Host, global.Env.Redis.Port),
		Password:     global.Env.Redis.Password,
		MinIdleConns: global.Env.Redis.MinIdle,
		PoolSize:     global.Env.Redis.Pool,
		IdleTimeout:  time.Duration(global.Env.Redis.Timeout),
	})
	err := redisCache.Ping().Err()
	if err != nil {
		panic(fmt.Sprintf("redis 初始化失败: %v", err))
	}
}

func GetRedis() *redis.Client {
	return redisCache
}
