package config

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

var mail *gomail.Dialer // 邮件执行对象

func GetMail() *gomail.Dialer {
	return mail
}

// InitializeMail 用于邮箱初始化
func InitializeMail() error {
	mail = gomail.NewDialer(GetConfig().Mail.Addr, GetConfig().Mail.Port, GetConfig().Mail.Username, GetConfig().Mail.Password)
	mail.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return nil
}
