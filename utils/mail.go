package utils

import (
	"airbox/config"

	"gopkg.in/gomail.v2"
)

// SendCaptcha send the captcha used to register and reset something
func SendCaptcha(email, captcha string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Env.Mail.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "来自AirBox的邮件提醒！")
	body := "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">\n" +
		"<html lang=\"en\">\n" +
		"<head>\n" +
		"    <meta charset=\"UTF-8\">\n" +
		"    <title>来自<a href=\"\">AirBox</a>的邮件提醒</title>\n" +
		"    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/>\n" +
		"</head>\n" +
		"<body style=\"padding: 0;\">\n" +
		"<table width=\"700\" border=\"0\" align=\"center\" cellspacing=\"0\" style=\"width:700px;\">\n" +
		"    <tbody>\n" +
		"    <tr>\n" +
		"        <td>\n" +
		"            <div style=\"width:680px;padding:0 10px;margin:0 auto;\">\n" +
		"                <div style=\"line-height:1.5;font-size:14px;margin-bottom:25px;color:#4d4d4d;\">\n" +
		"                    <strong style=\"display:block;margin-bottom:15px;\">\n" +
		"                        亲爱的用户：\n" +
		"                        <span style=\"color:rgb(129, 65, 134);font-size: 16px;\">" + email + "</span>\n" +
		"                    </strong>\n" +
		"                    <strong style=\"display:block;margin-bottom:15px;\">\n" +
		"                        您的验证码是：\n" +
		"                        <span style=\"color:#ff0000;font-size: 24px\">" + captcha + "</span>\n" +
		"                    </strong>\n" +
		"                </div>\n" +
		"                <div style=\"margin-bottom:30px;\">\n" +
		"                    <small style=\"display:block;margin-bottom:20px;font-size:12px;\">\n" +
		"                        <p style=\"color:#747474;\">\n" +
		"                            注意：此操作会泄露您的验证码。\n" +
		"                        </p>\n" +
		"                    </small>\n" +
		"                </div>\n" +
		"            </div>\n" +
		"        </td>\n" +
		"    </tr>\n" +
		"    </tbody>\n" +
		"</table>\n" +
		"</body>\n" +
		"</html>"
	m.SetBody("text/html", body)
	return config.Mail.DialAndSend(m)
}

// SendResetLink send the link used to reset password
func SendResetLink(email, link string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Env.Mail.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "来自AirBox的重置密码申请！")
	content := "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">\n" +
		"<html lang=\"en\">\n" +
		"<head>\n" +
		"    <meta charset=\"UTF-8\">\n" +
		"    <title>来自<a href=\"\">AirBox</a>的重置密码申请</title>\n" +
		"    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/>\n" +
		"</head>\n" +
		"<body style=\"padding: 0;\">\n" +
		"<table width=\"700\" border=\"0\" align=\"center\" cellspacing=\"0\" style=\"width:700px;\">\n" +
		"    <tbody>\n" +
		"    <tr>\n" +
		"        <td>\n" +
		"            <div style=\"width:680px;padding:0 10px;margin:0 auto;\">\n" +
		"                <div style=\"line-height:1.5;font-size:14px;margin-bottom:25px;color:#4d4d4d;\">\n" +
		"                    <strong style=\"display:block;margin-bottom:15px;\">\n" +
		"                        亲爱的用户：\n" +
		"                        <span style=\"color:rgb(129, 65, 134);font-size: 16px;\">" + email + "</span>\n" +
		"                    </strong>\n" +
		"                    <strong style=\"display:block;margin-bottom:15px;\">\n" +
		"                        您重置密码的链接是：\n" +
		"                        <span style=\"font-size: 12px\"><br>" +
		"                            <a style=\"color: #ff0000\" href=\"" + link + "\">" + link + "</a>" +
		"                        </span>\n" +
		"                    </strong>\n" +
		"                </div>\n" +
		"                <div style=\"margin-bottom:30px;\">\n" +
		"                    <small style=\"display:block;margin-bottom:20px;font-size:12px;\">\n" +
		"                        <p style=\"color:#747474;\">\n" +
		"                            注意：此操作会泄露您的密码。如非本人操作，请及时登录并修改密码以保证帐户安全\n" +
		"                        </p>\n" +
		"                    </small>\n" +
		"                </div>\n" +
		"            </div>\n" +
		"        </td>\n" +
		"    </tr>\n" +
		"    </tbody>\n" +
		"</table>\n" +
		"</body>\n" +
		"</html>"
	m.SetBody("text/html", content)
	return config.Mail.DialAndSend(m)
}
