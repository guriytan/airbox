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

func (c *RedisClient) GetToken(name string) string {
	return RedisCache.Get(KeyToken + name).Val()
}

func (*RedisClient) SetToken(name, token string) error {
	return RedisCache.Set(KeyToken+name, token, TokenUserExpiration).Err()
}
