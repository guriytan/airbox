package cache

import (
	"context"
	"sync"

	"airbox/global"

	"github.com/go-redis/redis"
)

type RedisClient struct {
	client *redis.Client
}

var (
	client     *RedisClient
	clientOnce sync.Once
)

func GetRedisClient(cache *redis.Client) *RedisClient {
	clientOnce.Do(func() {
		client = &RedisClient{client: cache}
	})
	return client
}

func (r *RedisClient) GetCaptcha(ctx context.Context, email string) (string, error) {
	return r.client.WithContext(ctx).Get(global.KeyCaptcha + email).Result()
}

func (r *RedisClient) DeleteCaptcha(ctx context.Context, email string) error {
	return r.client.WithContext(ctx).Del(global.KeyCaptcha + email).Err()
}

func (r *RedisClient) SetCaptcha(ctx context.Context, email, captcha string) error {
	return r.client.WithContext(ctx).Set(global.KeyCaptcha+email, captcha, global.TokenEmailExpiration).Err()
}

func (r *RedisClient) GetToken(ctx context.Context, name string) (string, error) {
	return r.client.WithContext(ctx).Get(global.KeyToken + name).Result()
}

func (r *RedisClient) SetToken(ctx context.Context, name, token string) error {
	return r.client.WithContext(ctx).Set(global.KeyToken+name, token, global.TokenUserExpiration).Err()
}
