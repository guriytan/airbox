package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reg, _ = regexp.Compile("\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*") // 邮箱匹配执行对象
)

// Epoch return current time in second.
func Epoch() int64 {
	return time.Now().Unix()
}

// Exp return expiration time in second
func Exp(ttl time.Duration) int64 {
	return time.Now().Add(ttl).Unix()
}

// CheckEmailFormat 检查邮箱格式
func CheckEmailFormat(email string) bool {
	return reg.MatchString(email)
}

// AddIndexToFilename used to rename the file if the file is exist
func AddIndexToFilename(file string, index int) string {
	split := strings.LastIndex(file, ".")
	return file[:split] + "(" + strconv.FormatInt(int64(index), 10) + ")" + file[split:]
}
