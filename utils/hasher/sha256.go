package hasher

import (
	"crypto/sha256"
	"encoding/hex"

	json "github.com/json-iterator/go"
)

type sha256Hash struct{}

func GetSha256() Hash {
	return &sha256Hash{}
}

func (s *sha256Hash) Hash(values ...interface{}) string {
	message, _ := json.Marshal(values)
	hash := sha256.New()
	hash.Write(message)
	return hex.EncodeToString(hash.Sum(nil))
}
