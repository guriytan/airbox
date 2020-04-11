package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

// SHA256Sum evaluates the sha256 of file
func SHA256Sum(filePath string) (result string, err error) {
	return sum(filePath, sha256.New())
}

// MD5Sum evaluates the md5 of file
func MD5Sum(filePath string) (result string, err error) {
	return sum(filePath, md5.New())
}

func sum(filePath string, h hash.Hash) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err = io.Copy(h, file); err != nil {
		return
	}

	result = hex.EncodeToString(h.Sum(nil))
	return
}
