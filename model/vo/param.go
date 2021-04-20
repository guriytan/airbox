package vo

import (
	"strconv"

	"airbox/global"
)

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
	FileID int64  `json:"file_id,string,omitempty"  form:"file_id"`
	Link   string `json:"link"                      form:"link"`
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
	FileID   int64           `json:"file_id,string,omitempty"   form:"file_id"   uri:"file_id"`
	Name     string          `json:"name"                       form:"name"`
	FatherID int64           `json:"father_id,string,omitempty" form:"father_id"`
	Size     int64           `json:"size"                       form:"size"`
	Hash     string          `json:"hash"                       form:"hash"`
	Type     global.FileType `json:"type"                       form:"type"`

	PageParam
}

type UpdateFileModel struct {
	FileID   int64                `json:"file_id,string,omitempty"   form:"file_id"`
	FatherID int64                `json:"father_id,string,omitempty" form:"father_id"`
	Name     string               `json:"name"                       form:"name"`
	OptType  global.OperationType `json:"opt_type"                   form:"opt_type"`
}

type TypeModel struct {
	FatherID string          `json:"father_id,omitempty" form:"father_id"`
	Type     global.FileType `json:"type" form:"type"`

	PageParam
}

func (t *TypeModel) GetFatherID() *int64 {
	fatherID, err := strconv.ParseInt(t.FatherID, 10, 64)
	if err != nil {
		return nil
	}
	return &fatherID
}

type PageParam struct {
	Cursor int64 `json:"cursor,string,omitempty"    form:"cursor"`
	Limit  int   `json:"limit"                      form:"limit"`
}
