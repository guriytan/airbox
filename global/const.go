package global

import (
	"os"
	"time"
)

const (
	KeyRequestID = "request_id"
	KeyFunction  = "function"
	KeyUserID    = "user_id"
	KeyIP        = "ip"
)

// 文件类型
type FileType int

const (
	FileFolderType = iota
	FileMusicType
	FileVideoType
	FileDocumentType
	FilePictureType
	FileOtherType
)

const (
	DefaultFatherID = "air_box_default_father_id"
)

// 用户名及密码长度限制
const (
	UserMaxLength = 18
	UserMinLength = 6
)

// 常用错误状态返回信息
const (
	ErrorOutOfDated         = "登录信息已过期"
	ErrorWithoutToken       = "用户未登录"
	ErrorSSO                = "用户已在其他设备登录"
	ErrorOutOfSpace         = "空间不足"
	ErrorOfCaptcha          = "验证码错误"
	ErrorOfConflict         = "该文件夹下已存在同名文件或文件夹"
	ErrorOfUsername         = "用户名长度应在6位至18位之间"
	ErrorOfPassword         = "密码长度应在6位至18位之间"
	ErrorOfEmail            = "邮箱不能为空或格式错误"
	ErrorOfSystem           = "系统错误"
	ErrorOfRequestParameter = "请求参数错误"
	ErrorOfWithoutName      = "缺少名字"
	ErrorOfWrongToken       = "token错误"
	ErrorOfFileID           = "缺少文件id"
	ErrorOfForbidRegister   = "禁止注册"
	ErrorOfExistUsername    = "用户已存在"
	ErrorOfExistEmail       = "邮箱已存在"
	ErrorOfWrongPassword    = "密码错误"
	ErrorOfSameEmail        = "新邮箱不能与原邮箱一致"
	ErrorOfSamePassword     = "不能与原密码一致"
	ErrorOfExpectedLink     = "链接失效"
	ErrorOfCopyFile         = "不能复制或移动到自身"
	ErrorDownloadFile       = "下载文件失败"
)

// 文件上传相关配置
const (
	FileTempSuffix = ".temp"           // Temp file suffix
	FilePermMode   = os.FileMode(0666) // Default file permission
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
	KeyToken    = "KEY_REDIS_TOKEN" + KeyInternal
)

// token密钥
var (
	SecretKeyUser  = []byte("TOq89fQY4tp29J4g") // The key required for the user token
	SecretKeyEmail = []byte("12H4ywQr8f2hD023") // The key required to reset password
	SecretKeyFile  = []byte("049fhAwf592hOc42") // The key required to share file
)
