package vo

import "airbox/global"

type LoginModel struct {
	UserKey  string `json:"user_key" form:"user_key"`
	Password string `json:"password" form:"password"`
}

type UserModel struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email"    form:"email"`
	Password string `json:"password" form:"password"`
	Code     string `json:"code"     form:"code"`
}

type ShareModel struct {
	FileID string `json:"file_id" form:"file_id"`
	Link   string `json:"link"    form:"link"`
}

type TokenModel struct {
	Token string `json:"token" form:"token"`
}

type ResetPasswordModel struct {
	Origin string `json:"origin" form:"origin"`
	Reset  string `json:"reset"  form:"reset"`
}

type ResetEmailModel struct {
	Email string `json:"email" form:"email"`
	Code  string `json:"code"  form:"code"`
}

type FileModel struct {
	FileID   string          `json:"file_id"   form:"file_id"   uri:"file_id"`
	Name     string          `json:"name"      form:"name"`
	FatherID string          `json:"father_id" form:"father_id"`
	Size     int64           `json:"size"      form:"size"`
	Hash     string          `json:"hash"      form:"hash"`
	Type     global.FileType `json:"type"      form:"type"`
}

type UpdateFileModel struct {
	FileID   string               `json:"file_id"   form:"file_id"`
	FatherID string               `json:"father_id" form:"father_id"`
	Name     string               `json:"name"      form:"name"`
	OptType  global.OperationType `json:"opt_type"  form:"opt_type"`
}

type TypeModel struct {
	FatherID string          `json:"father_id" form:"father_id"`
	Type     global.FileType `json:"type"      form:"type"`
}
