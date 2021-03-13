package encryption

import (
	json "github.com/json-iterator/go"

	"airbox/global"
)

// GenerateUserToken return the token of user which used to verify the authority
func GenerateUserToken(key ...interface{}) (string, error) {
	marshal, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	return aesEncryption(string(marshal), exp(global.TokenUserExpiration), global.SecretKeyUser)
}

// ParseUserToken return the struct of user by parsing the user token
func ParseUserToken(token string, value interface{}) (int64, error) {
	content, exp, err := aesDecryption(token, global.SecretKeyUser)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal([]byte(content), value)
	return exp, err
}

// GenerateEmailToken return the token of email which used to reset the password
func GenerateEmailToken(email string) (string, error) {
	return aesEncryption(email, exp(global.TokenEmailExpiration), global.SecretKeyEmail)
}

// ParseEmailToken return the  email and the time
func ParseEmailToken(token string) (string, int64, error) {
	email, exp, err := aesDecryption(token, global.SecretKeyEmail)
	if err != nil {
		return "", 0, err
	}
	return email, exp, nil
}

// GenerateShareToken return the token of link which can download file with no authority
func GenerateShareToken(id string) (string, error) {
	return aesEncryption(id, exp(global.TokenFileExpiration), global.SecretKeyFile)
}

// ParseShareToken return the file id and the time
func ParseShareToken(token string) (string, int64, error) {
	email, exp, err := aesDecryption(token, global.SecretKeyFile)
	if err != nil {
		return "", 0, err
	}
	return email, exp, nil
}
