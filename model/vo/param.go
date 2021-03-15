package vo

import "airbox/global"

type LoginModel struct {
	UserKey  string `json:"user_key"`
	Password string `json:"password"`
}

type UserModel struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type ShareModel struct {
	FileID string `json:"file_id"`
	Link   string `json:"link"`
}

type TokenModel struct {
	Token string `json:"token"`
}

type ResetPasswordModel struct {
	Origin string `json:"origin"`
	Reset  string `json:"reset"`
}

type ResetEmailModel struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type FileModel struct {
	FileID   string `json:"file_id" form:"file_id" uri:"file_id"`
	Name     string `json:"name" form:"name"`
	FatherID string `json:"father_id" form:"father_id"`
	Size     int64  `json:"size" form:"size"`
	Hash     string `json:"hash" form:"hash"`
}

type UpdateFileModel struct {
	FileID   string               `json:"file_id" form:"file_id"`
	FatherID string               `json:"father_id" form:"father_id"`
	Name     string               `json:"name" form:"name"`
	OptType  global.OperationType `json:"opt_type" form:"opt_type"`
}

type TypeModel struct {
	Type global.FileType `json:"type"`
}
