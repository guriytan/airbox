package utils

import (
	"crypto/sha256"
	"encoding/hex"

	json "github.com/json-iterator/go"
)

func Hash(values ...interface{}) string {
	message, _ := json.Marshal(values)
	hash := sha256.New()
	hash.Write(message)
	return hex.EncodeToString(hash.Sum(nil))
}
