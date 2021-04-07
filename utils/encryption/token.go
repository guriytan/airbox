package encryption

import (
	"strconv"

	json "github.com/json-iterator/go"

	"airbox/global"
)

// GenerateUserToken return the token of user which used to verify the authority
func GenerateUserToken(key interface{}) (string, error) {
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

// GenerateEmailToken return the token of user id which used to reset the password
func GenerateEmailToken(id int64) (string, error) {
	return aesEncryption(strconv.FormatInt(id, 10), exp(global.TokenEmailExpiration), global.SecretKeyEmail)
}

// ParseEmailToken return the user id and the time
func ParseEmailToken(token string) (int64, int64, error) {
	userID, exp, err := aesDecryption(token, global.SecretKeyEmail)
	if err != nil {
		return 0, 0, err
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return id, exp, nil
}

// GenerateShareToken return the token of link which can download file with no authority
func GenerateShareToken(fileID int64) (string, error) {
	return aesEncryption(strconv.FormatInt(fileID, 10), exp(global.TokenFileExpiration), global.SecretKeyFile)
}

// ParseShareToken return the file id and the time
func ParseShareToken(token string) (int64, int64, error) {
	fileID, exp, err := aesDecryption(token, global.SecretKeyFile)
	if err != nil {
		return 0, 0, err
	}
	id, err := strconv.ParseInt(fileID, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return id, exp, nil
}
