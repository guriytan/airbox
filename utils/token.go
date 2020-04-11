package utils

import (
	"airbox/config"
	"airbox/model"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"math/rand"
	"strconv"
	"strings"

	json "github.com/json-iterator/go"
)

const (
	// CaptchaNumber iota means the chapcha has three type of char
	// CaptchaNumber is the number
	CaptchaNumber int = iota
	// CaptchaUppercase is the uppercase of letter
	CaptchaUppercase
	// CaptchaLowercase is the lowwercase of letter
	CaptchaLowercase

	// CaptchaLength indicate the Length of captcha
	CaptchaLength = 8
	// CaptchaType means captcha has three type, Number, Uppercase, Lowercase
	CaptchaType = 3
)

// GetEmailCaptcha provides to the register, update email and reset password
func GetEmailCaptcha() string {
	captcha := make([]byte, CaptchaLength)
	for i := 0; i < CaptchaLength; i++ {
		switch rand.Intn(CaptchaType) {
		case CaptchaNumber:
			captcha[i] = byte(rand.Intn(10) + 48)
		case CaptchaUppercase:
			captcha[i] = byte(rand.Intn(26) + 65)
		case CaptchaLowercase:
			captcha[i] = byte(rand.Intn(26) + 97)
		}
	}
	return string(captcha)
}

// GenerateUserToken return the token of user which used to verify the authority
func GenerateUserToken(user *model.User) (string, error) {
	marshal, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return aesEncryption(string(marshal), Exp(config.TokenUserExpiration), config.SecretKeyUser)
}

// ParseUserToken return the struct of user by parsing the user token
func ParseUserToken(token string) (*model.User, int64, error) {
	content, exp, err := aesDecryption(token, config.SecretKeyUser)
	if err != nil {
		return nil, 0, err
	}
	user := &model.User{}
	err = json.Unmarshal([]byte(content), user)
	return user, exp, err
}

// GenerateEmailToken return the token of email which used to reset the password
func GenerateEmailToken(email string) (string, error) {
	return aesEncryption(email, Exp(config.TokenEmailExpiration), config.SecretKeyEmail)
}

// ParseEmailToken return the  email and the time
func ParseEmailToken(token string) (string, int64, error) {
	email, exp, err := aesDecryption(token, config.SecretKeyEmail)
	if err != nil {
		return "", 0, err
	}
	return email, exp, nil
}

// GenerateShareToken return the token of link which can download file with no authority
func GenerateShareToken(id string) (string, error) {
	return aesEncryption(id, Exp(config.TokenFileExpiration), config.SecretKeyFile)
}

// ParseShareToken return the file id and the time
func ParseShareToken(token string) (string, int64, error) {
	email, exp, err := aesDecryption(token, config.SecretKeyFile)
	if err != nil {
		return "", 0, err
	}
	return email, exp, nil
}

// aesEncryption 为使用AES方法的加密方法
func aesEncryption(content string, exp int64, key []byte) (string, error) {
	origin := []byte(content + config.KeyInternal + strconv.FormatInt(exp, 10))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	padding := blockSize - len(origin)%blockSize
	origin = append(origin, bytes.Repeat([]byte{byte(padding)}, padding)...)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypt := make([]byte, len(origin))
	blockMode.CryptBlocks(encrypt, origin)
	return base64.RawURLEncoding.EncodeToString(encrypt), nil
}

// aesDecryption 为使用AES方法的解密方法
func aesDecryption(crypto string, key []byte) (content string, exp int64, err error) {
	encrypt, err := base64.RawURLEncoding.DecodeString(crypto)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	decrypt := make([]byte, len(encrypt))
	blockMode.CryptBlocks(decrypt, encrypt)
	length := len(decrypt)
	decrypt = decrypt[:(length - int(decrypt[length-1]))]
	para := strings.Split(string(decrypt), config.KeyInternal)
	if len(para) == 2 {
		exp, err = strconv.ParseInt(para[1], 10, 64)
		return para[0], exp, err
	}
	return "", 0, nil
}
