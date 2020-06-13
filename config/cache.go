package config

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

var redisCache *redis.Client

// InitializeRedis 用于Redis初始化
func InitializeRedis() {
	redisCache = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", Env.Redis.Host, Env.Redis.Port),
		Password:     Env.Redis.Password,
		MinIdleConns: Env.Redis.MinIdle,
		PoolSize:     Env.Redis.Pool,
		IdleTimeout:  time.Duration(Env.Redis.Timeout),
	})
	err := redisCache.Ping().Err()
	if err != nil {
		panic(fmt.Sprintf("redis 初始化失败: %v", err))
	}
}

func GetRedis() *redis.Client {
	return redisCache
}
