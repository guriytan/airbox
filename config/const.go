package config

import (
	"os"
	"time"
)

// 文件类型
type FileType int

const (
	FileMusicType = iota
	FileVideoType
	FileDocumentType
	FilePictureType
	FileOtherType
)

// 用户名及密码长度限制
const (
	UserMaxLength = 18
	UserMinLength = 6
)

// 常用错误状态返回信息
const (
	ErrorOutOfDated   = "token is outdated"
	ErrorWithoutToken = "token is needed"
	ErrorOutOfSpace   = "空间不足"
	ErrorOfCaptcha    = "验证码错误"
	ErrorOfConflict   = "该文件夹下已存在同名文件或文件夹"
)

// 文件上传相关配置
const (
	FileTempSuffix            = ".temp"           // Temp file suffix
	FilePermMode              = os.FileMode(0666) // Default file permission
	FileDownloadPartSize      = 100 * 1024        // Default download part size, 100KB
	FileGoroutine             = 4                 // Default routine, 4
	FileTimeout               = 30                // Timeout for transmission, 30 second
	FilePrefixMasterDirectory = "./store/"
)

// token及code的过期时间
const (
	TokenUserExpiration  = 6 * time.Hour    // User Token expiration Period, represented in second
	TokenEmailExpiration = 10 * time.Minute // Email Token expiration Period, represented in second
	TokenFileExpiration  = 24 * time.Hour   // File Token expiration Period, represented in second
)

// key配置
const (
	KeyInternal = "::"
	KeyCaptcha  = "KEY_REDIS_CAPTCHA" + KeyInternal
)

// token密钥
var (
	SecretKeyUser  = []byte("TOq89fQY4tp29J4g") // The key required for the user token
	SecretKeyEmail = []byte("12H4ywQr8f2hD023") // The key required to reset password
	SecretKeyFile  = []byte("049fhAwf592hOc42") // The key required to share file
)
