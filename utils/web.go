package utils

import (
	"regexp"
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
