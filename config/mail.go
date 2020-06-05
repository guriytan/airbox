package config

import (
	"airbox/global"
	"crypto/tls"
	"gopkg.in/gomail.v2"
)

var mail *gomail.Dialer // 邮件执行对象

func GetMail() *gomail.Dialer {
	return mail
}

// InitializeMail 用于邮箱初始化
func InitializeMail() {
	mail = gomail.NewDialer(global.Env.Mail.Addr, global.Env.Mail.Port, global.Env.Mail.Username, global.Env.Mail.Password)
	mail.TLSConfig = &tls.Config{InsecureSkipVerify: true}
}
