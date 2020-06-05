package service

import (
	"airbox/cache"
	"airbox/global"
	"airbox/utils"
	"airbox/utils/encryption"
	"github.com/pkg/errors"
)

type AuthService struct {
	redis *cache.RedisClient
}

var auth *AuthService

func GetAuthService() *AuthService {
	if auth == nil {
		auth = &AuthService{
			redis: cache.GetRedisClient(),
		}
	}
	return auth
}

// VerifyEmailCaptcha 从缓存中读取key为email的值与code判断是否一致
// 当相等时返回true，不相等返回false
func (c *AuthService) VerifyEmailCaptcha(email string, code string) bool {
	if captcha := c.redis.GetCaptcha(email); captcha == code {
		return true
	}
	return false
}

// SendCaptcha 生成随机验证码发送至邮箱
func (c *AuthService) SendCaptcha(email string) error {
	captcha := encryption.GetEmailCaptcha()
	if err := c.redis.SetCaptcha(email, captcha); err != nil {
		return errors.WithStack(err)
	}
	if err := utils.SendCaptcha(email, captcha); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteCaptcha 删除邮箱验证码
func (c *AuthService) DeleteCaptcha(key string) {
	c.redis.DeleteCaptcha(key)
}

// SendResetLink 根据邮箱生成链接发送至邮箱
func (c *AuthService) SendResetLink(id, email string) error {
	captcha, err := encryption.GenerateEmailToken(id)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := utils.SendResetLink(email, global.Env.Web.Site+"/reset/"+captcha); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// VerifyToken 验证请求里的token和redis中的token是否一致
func (c *AuthService) VerifyToken(name, token string) bool {
	return c.redis.GetToken(name) == token
}

// SetToken 储存当前最新的token
func (c *AuthService) SetToken(name, token string) error {
	if err := c.redis.SetToken(name, token); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
