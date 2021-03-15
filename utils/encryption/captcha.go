package encryption

import (
	"math/rand"
	"time"
)

const (
	// CaptchaNumber iota means the Captcha has three type of char
	// CaptchaNumber is the number
	CaptchaNumber int = iota
	// CaptchaUppercase is the uppercase of letter
	CaptchaUppercase
	// CaptchaLowercase is the lowercase of letter
	CaptchaLowercase
	// CaptchaLength indicate the Length of captcha
	CaptchaLength = 8
	// CaptchaType means captcha has three type, Number, Uppercase, Lowercase
	CaptchaType = 3
)

// GetEmailCaptcha provides to the register, update email and reset password
func GetEmailCaptcha() string {
	var source = rand.New(rand.NewSource(time.Now().UnixNano()))
	captcha := make([]byte, CaptchaLength)
	for i := 0; i < CaptchaLength; i++ {
		switch source.Intn(CaptchaType) {
		case CaptchaNumber:
			captcha[i] = byte(source.Intn(10) + 48)
		case CaptchaUppercase:
			captcha[i] = byte(source.Intn(26) + 65)
		case CaptchaLowercase:
			captcha[i] = byte(source.Intn(26) + 97)
		}
	}
	return string(captcha)
}
