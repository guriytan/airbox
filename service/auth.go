package service

import (
	"airbox/cache"
	"airbox/config"
	"airbox/utils"
)

type CaptchaService struct {
	redis *cache.RedisClient
}

var captcha *CaptchaService

func GetCaptchaService() *CaptchaService {
	if captcha == nil {
		captcha = &CaptchaService{
			redis: cache.GetRedisClient(),
		}
	}
	return captcha
}

// VerifyEmailCaptcha 从缓存中读取key为email的值与code判断是否一致
// 当相等时返回true，不相等返回false
func (c *CaptchaService) VerifyEmailCaptcha(email string, code string) bool {
	if captcha := c.redis.GetCaptcha(email); captcha == code {
		return true
	}
	return false
}

// SendCaptcha 生成随机验证码发送至邮箱
func (c *CaptchaService) SendCaptcha(email string) error {
	captcha := utils.GetEmailCaptcha()
	if err := c.redis.SetCaptcha(email, captcha); err != nil {
		return err
	}
	return utils.SendCaptcha(email, captcha)
}

// SendResetLink 根据邮箱生成链接发送至邮箱
func (c *CaptchaService) SendResetLink(id, email string) error {
	captcha, err := utils.GenerateEmailToken(id)
	if err != nil {
		return err
	}
	return utils.SendResetLink(email, config.Env.Web.Site+"/reset/"+captcha)
}
