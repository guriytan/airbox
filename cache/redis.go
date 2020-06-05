package cache

import (
	"airbox/global"
	"github.com/go-redis/redis"
)

type RedisClient struct {
	*redis.Client
}

var client *RedisClient

func GetRedisClient() *RedisClient {
	if client == nil {
		client = &RedisClient{
			global.REDIS,
		}
	}
	return client
}

func (r *RedisClient) GetCaptcha(email string) string {
	return r.Get(global.KeyCaptcha + email).Val()
}

func (r *RedisClient) DeleteCaptcha(email string) {
	r.Del(global.KeyCaptcha + email)
}

func (r *RedisClient) SetCaptcha(email, captcha string) error {
	return r.Set(global.KeyCaptcha+email, captcha, global.TokenEmailExpiration).Err()
}

func (r *RedisClient) GetToken(name string) string {
	return r.Get(global.KeyToken + name).Val()
}

func (r *RedisClient) SetToken(name, token string) error {
	return r.Set(global.KeyToken+name, token, global.TokenUserExpiration).Err()
}
