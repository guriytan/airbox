package service

import (
	"context"
	"sync"

	"airbox/cache"
	"airbox/config"
	"airbox/utils"
	"airbox/utils/encryption"
)

type AuthService struct {
	redis *cache.RedisClient
}

var (
	auth     *AuthService
	authOnce sync.Once
)

func GetAuthService() *AuthService {
	authOnce.Do(func() {
		auth = &AuthService{redis: cache.GetRedisClient(config.GetCache())}
	})
	return auth
}

// VerifyEmailCaptcha 从缓存中读取key为email的值与code判断是否一致
// 当相等时返回true，不相等返回false
func (c *AuthService) VerifyEmailCaptcha(ctx context.Context, email string, code string) bool {
	if captcha := c.redis.GetCaptcha(ctx, email); captcha == code {
		return true
	}
	return false
}

// SendCaptcha 生成随机验证码发送至邮箱
func (c *AuthService) SendCaptcha(ctx context.Context, email string) error {
	captcha := encryption.GetEmailCaptcha()
	if err := c.redis.SetCaptcha(ctx, email, captcha); err != nil {
		return err
	}
	if err := utils.SendCaptcha(email, captcha); err != nil {
		return err
	}
	return nil
}

// DeleteCaptcha 删除邮箱验证码
func (c *AuthService) DeleteCaptcha(ctx context.Context, key string) {
	c.redis.DeleteCaptcha(ctx, key)
}

// SendResetLink 根据邮箱生成链接发送至邮箱
func (c *AuthService) SendResetLink(id, email string) error {
	captcha, err := encryption.GenerateEmailToken(id)
	if err != nil {
		return err
	}
	if err := utils.SendResetLink(email, config.GetConfig().Web.Site+"/reset/"+captcha); err != nil {
		return err
	}
	return nil
}

// VerifyToken 验证请求里的token和redis中的token是否一致
func (c *AuthService) VerifyToken(ctx context.Context, name, token string) bool {
	return c.redis.GetToken(ctx, name) == token
}

// SetToken 储存当前最新的token
func (c *AuthService) SetToken(ctx context.Context, name, token string) error {
	if err := c.redis.SetToken(ctx, name, token); err != nil {
		return err
	}
	return nil
}
