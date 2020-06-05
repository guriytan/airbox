package encryption

import (
	"airbox/global"
	"airbox/model"
	"airbox/utils"
	json "github.com/json-iterator/go"
)

// GenerateUserToken return the token of user which used to verify the authority
func GenerateUserToken(user *model.User) (string, error) {
	marshal, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return aesEncryption(string(marshal), utils.Exp(global.TokenUserExpiration), global.SecretKeyUser)
}

// ParseUserToken return the struct of user by parsing the user token
func ParseUserToken(token string) (*model.User, int64, error) {
	content, exp, err := aesDecryption(token, global.SecretKeyUser)
	if err != nil {
		return nil, 0, err
	}
	user := &model.User{}
	err = json.Unmarshal([]byte(content), user)
	return user, exp, err
}

// GenerateEmailToken return the token of email which used to reset the password
func GenerateEmailToken(email string) (string, error) {
	return aesEncryption(email, utils.Exp(global.TokenEmailExpiration), global.SecretKeyEmail)
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
	return aesEncryption(id, utils.Exp(global.TokenFileExpiration), global.SecretKeyFile)
}

// ParseShareToken return the file id and the time
func ParseShareToken(token string) (string, int64, error) {
	email, exp, err := aesDecryption(token, global.SecretKeyFile)
	if err != nil {
		return "", 0, err
	}
	return email, exp, nil
}
