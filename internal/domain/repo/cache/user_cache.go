// 保留接口定义和常量
package cache

import (
	"context"
	"time"

	"GochatIM/internal/domain/entity"
)

const (
	// 用户信息缓存前缀
	UserCachePrefix = "user:"
	// 用户在线状态前缀
	UserOnlinePrefix = "user_online:"
	// 用户token前缀
	UserTokenPrefix = "user_token:"
)

// UserCache 用户缓存接口
type UserCache interface {
	// 用户信息缓存
	SetUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, userID uint64) (*entity.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
	
	// Token相关
	SetUserToken(ctx context.Context, userID uint64, tokenID string, token string, expiration time.Duration) error
	GetUserToken(ctx context.Context, userID uint64, tokenID string) (string, error)
	RevokeUserToken(ctx context.Context, userID uint64, tokenID string) error
	RevokeAllUserTokens(ctx context.Context, userID uint64) error
	
	// 在线状态
	SetUserOnline(ctx context.Context, userID uint64, deviceID string) error
	SetUserOffline(ctx context.Context, userID uint64, deviceID string) error
	IsUserOnline(ctx context.Context, userID uint64) (bool, error)
}
