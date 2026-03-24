package repository

import (
	"Hai-Service/domain"
	"context"
)

// UserRepository 定义了用户仓储需要实现的接口
type UserRepository interface {
	// Save 保存用户信息（创建或更新）
	Save(ctx context.Context, user *domain.User) error

	// FindByUserID 通过业务ID查找用户
	FindByUserID(ctx context.Context, userID string) (*domain.User, error)

	// FindByID 通过数据库主键查找用户
	FindByID(ctx context.Context, id uint) (*domain.User, error)
}
