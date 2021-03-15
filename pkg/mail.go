package pkg

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"

	"airbox/config"
)

var mail *gomail.Dialer // 邮件执行对象

func GetMail() *gomail.Dialer {
	return mail
}

// InitializeMail 用于邮箱初始化
func InitializeMail() error {
	mail = gomail.NewDialer(config.GetConfig().Mail.Addr, config.GetConfig().Mail.Port, config.GetConfig().Mail.Username, config.GetConfig().Mail.Password)
	mail.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return nil
}
