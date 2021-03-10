package base

import (
	"context"

	"airbox/model"

	"gorm.io/gorm"
)

// 用户数据库操作接口
type UserDao interface {
	InsertUser(ctx context.Context, tx *gorm.DB, user *model.User) error

	DeleteUserByID(ctx context.Context, tx *gorm.DB, id string) error

	UpdateUser(ctx context.Context, user *model.User) error

	SelectUserByID(ctx context.Context, id string) (*model.User, error)
	SelectUserByPwdAndNameOrEmail(ctx context.Context, name, pwd string) (*model.User, error)
	SelectUserByName(ctx context.Context, username string) (*model.User, error)
	SelectUserByEmail(ctx context.Context, email string) (*model.User, error)
}
