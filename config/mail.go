package config

import (
	"crypto/tls"
	"gopkg.in/gomail.v2"
)

var Mail *gomail.Dialer // 邮件执行对象

// initializeMail 用于邮箱初始化
func initializeMail() {
	Mail = gomail.NewDialer(Env.Mail.Addr, Env.Mail.Port, Env.Mail.Username, Env.Mail.Password)
	Mail.TLSConfig = &tls.Config{InsecureSkipVerify: true}
}
