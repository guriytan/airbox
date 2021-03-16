package base

import (
	"context"

	"airbox/model/do"

	"gorm.io/gorm"
)

// 用户数据库操作接口
type UserDao interface {
	InsertUser(ctx context.Context, tx *gorm.DB, user *do.User) error

	DeleteUserByID(ctx context.Context, tx *gorm.DB, userID string) error

	UpdateUser(ctx context.Context, user *do.User) error

	SelectUserByID(ctx context.Context, userID string) (*do.User, error)
	SelectUserByPwdAndNameOrEmail(ctx context.Context, name, pwd string) (*do.User, error)
	SelectUserByName(ctx context.Context, username string) (*do.User, error)
	SelectUserByEmail(ctx context.Context, email string) (*do.User, error)
}
