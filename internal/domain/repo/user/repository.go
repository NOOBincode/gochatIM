package user

import (
	"context"
	"GochatIM/internal/domain/entity"
)

// Repository 用户仓储接口
type Repository interface {
	// FindByID 根据ID查找用户
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	
	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	
	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	
	// FindByPhone 根据手机号查找用户
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)
	
	// FindByIDs 批量查询用户
	FindByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)
	
	// Create 创建用户
	Create(ctx context.Context, user *entity.User) error
	
	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error
	
	// Delete 删除用户
	Delete(ctx context.Context, id uint64) error
}