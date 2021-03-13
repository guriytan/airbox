package service

import (
	"context"
	"sync"

	"airbox/cache"
	"airbox/config"
	"airbox/logger"
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
	log := logger.GetLogger(ctx, "VerifyEmailCaptcha")
	captcha, err := c.redis.GetCaptcha(ctx, email)
	if err != nil {
		log.WithError(err).Warnf("get captcha of email: %v failed", email)
		return false
	}
	log.Infof("check email: %v, captcha: %v, actual: %v", email, code, captcha)
	return captcha == code
}

// SendCaptcha 生成随机验证码发送至邮箱
func (c *AuthService) SendCaptcha(ctx context.Context, email string) error {
	log := logger.GetLogger(ctx, "SendCaptcha")
	captcha := encryption.GetEmailCaptcha()
	if err := c.redis.SetCaptcha(ctx, email, captcha); err != nil {
		log.WithError(err).Warnf("set captcha: %v, email: %v to redis failed", captcha, email)
		return err
	}
	if err := utils.SendCaptcha(email, captcha); err != nil {
		log.WithError(err).Warnf("send captcha: %v to email: %v failed", captcha, email)
		return err
	}
	log.Infof("send captcha: %v to email: %v success", captcha, email)
	return nil
}

// DeleteCaptcha 删除邮箱验证码
func (c *AuthService) DeleteCaptcha(ctx context.Context, key string) {
	log := logger.GetLogger(ctx, "DeleteCaptcha")
	if err := c.redis.DeleteCaptcha(ctx, key); err != nil {
		log.WithError(err).Warnf("delete captcha: %v failed", key)
	}
}

// SendResetLink 根据邮箱生成链接发送至邮箱
func (c *AuthService) SendResetLink(ctx context.Context, id, email string) error {
	log := logger.GetLogger(ctx, "SendResetLink")
	captcha, err := encryption.GenerateEmailToken(id)
	if err != nil {
		log.WithError(err).Warnf("generate email token: %v failed", id)
		return err
	}
	link := config.GetConfig().Web.Site + "/reset/" + captcha
	if err := utils.SendResetLink(email, link); err != nil {
		log.WithError(err).Warnf("send token: %v to email: %v failed", captcha, email)
		return err
	}
	log.Infof("send link: %v to email: %v success", link, email)
	return nil
}

// VerifyToken 验证请求里的token和redis中的token是否一致
func (c *AuthService) VerifyToken(ctx context.Context, name, token string) bool {
	log := logger.GetLogger(ctx, "VerifyToken")
	getToken, err := c.redis.GetToken(ctx, name)
	if err != nil {
		log.WithError(err).Warnf("get token of name: %v failed", name)
	}
	return getToken == token
}

// SetToken 储存当前最新的token
func (c *AuthService) SetToken(ctx context.Context, name, token string) error {
	log := logger.GetLogger(ctx, "SetToken")
	if err := c.redis.SetToken(ctx, name, token); err != nil {
		log.WithError(err).Warnf("set token: %v of name: %v failed", token, name)
		return err
	}
	return nil
}
