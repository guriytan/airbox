package config

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"
)

var (
	RedisCache *redis.Client
)

// initializeRedis 用于Redis初始化
func initializeRedis() {
	RedisCache = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", Env.Redis.Host, Env.Redis.Port),
		Password:     Env.Redis.Password,
		MinIdleConns: Env.Redis.MinIdle,
		PoolSize:     Env.Redis.Pool,
		IdleTimeout:  time.Duration(Env.Redis.Timeout),
	})
	err := RedisCache.Ping().Err()
	if err != nil {
		log.Fatalf("failed to connect redis: %s\n", err.Error())
	}
}
