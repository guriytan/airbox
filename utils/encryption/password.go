package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// EncryptPassword 用于加密用户密码
func EncryptPassword(password string) (result string) {
	bytes := []byte(password)
	sum := hmac.New(sha256.New, sha256.New().Sum(bytes)).Sum(bytes)
	return base64.StdEncoding.EncodeToString(sum)
}
