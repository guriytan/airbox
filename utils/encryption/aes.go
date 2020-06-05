package encryption

import (
	"airbox/global"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strconv"
	"strings"
)

// aesEncryption 为使用AES方法的加密方法
func aesEncryption(content string, exp int64, key []byte) (string, error) {
	origin := []byte(content + global.KeyInternal + strconv.FormatInt(exp, 10))
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
	para := strings.Split(string(decrypt), global.KeyInternal)
	if len(para) == 2 {
		exp, err = strconv.ParseInt(para[1], 10, 64)
		return para[0], exp, err
	}
	return "", 0, nil
}
