package hasher

import (
	"crypto/md5"
	"encoding/hex"

	json "github.com/json-iterator/go"
)

type md5Hash struct{}

func GetMD5() Hash {
	return &md5Hash{}
}

func (m *md5Hash) Hash(values ...interface{}) string {
	message, _ := json.Marshal(values)
	hash := md5.New()
	hash.Write(message)
	return hex.EncodeToString(hash.Sum(nil))
}
