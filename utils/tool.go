package utils

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	reg, _ = regexp.Compile("\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*") // 邮箱匹配执行对象
)

// Epoch return current time in second.
func Epoch() int64 {
	return time.Now().Unix()
}

// CheckEmailFormat 检查邮箱格式
func CheckEmailFormat(email string) bool {
	return reg.MatchString(email)
}

// AddSuffixToFilename used to rename the file if the file is exist
func AddSuffixToFilename(file string) string {
	split := strings.LastIndex(file, ".")
	return file[:split] + "-" + uuid.New().URN() + file[split:]
}
