package vo

type LoginModel struct {
	UserKey  string `json:"user_key"`
	Password string `json:"password"`
}

type UserModel struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type FileModel struct {
	FileID   string `json:"file_id"`
	Name     string `json:"name"`
	FatherID string `json:"father_id"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
	Type     string `json:"type"`
	Link     string `json:"link"`
}
