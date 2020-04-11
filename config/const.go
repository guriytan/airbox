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

const (
	UserMaxLength = 18
	UserMinLength = 6
)

const (
	CodeErrorOfRequest   = 400
	CodeErrorOfServer    = 500
	CodeErrorOfAuthority = 403
	CodeSuccess          = 200
)

const (
	FileTempSuffix       = ".temp"           // Temp file suffix
	FilePermMode         = os.FileMode(0666) // Default file permission
	FileDownloadPartSize = 100 * 1024        // Default download part size, 100KB
	FileGoroutine        = 4                 // Default routine, 4
	FileTimeout          = 30                // Timeout for transmission, 30 second
)

const PrefixMasterDirectory = "./store/"

const (
	TokenUserExpiration  = 6 * time.Hour    // User Token expiration Period, represented in second
	TokenEmailExpiration = 10 * time.Minute // Email Token expiration Period, represented in second
	TokenFileExpiration  = 24 * time.Hour   // File Token expiration Period, represented in second
)

const (
	KeyInternal = "::"
	KeyCaptcha  = "KEY_REDIS_CAPTCHA" + KeyInternal
)

var (
	SecretKeyUser  = []byte("TOq89fQY4tp29J4g") // The key required for the user token
	SecretKeyEmail = []byte("12H4ywQr8f2hD023") // The key required to reset password
	SecretKeyFile  = []byte("049fhAwf592hOc42") // The key required to share file
)
