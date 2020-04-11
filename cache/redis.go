package cache

import (
	. "airbox/config"
)

type RedisClient struct {
}

var redis *RedisClient

func GetRedisClient() *RedisClient {
	if redis == nil {
		redis = &RedisClient{}
	}
	return redis
}

func (*RedisClient) GetCaptcha(email string) string {
	return RedisCache.Get(KeyCaptcha + email).Val()
}

func (*RedisClient) DeleteCaptcha(email string) {
	RedisCache.Del(KeyCaptcha + email)
}

func (*RedisClient) SetCaptcha(email, captcha string) error {
	return RedisCache.Set(KeyCaptcha+email, captcha, TokenEmailExpiration).Err()
}
