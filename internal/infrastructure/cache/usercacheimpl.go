package cache

import (
	"context"
	"fmt"
	"time"

	"GochatIM/internal/domain/entity"
	"GochatIM/pkg/logger"
	"GochatIM/pkg/redis"

	"github.com/bytedance/sonic"
)

const (
	// 用户信息缓存前缀
	UserCachePrefix = "user:"
	// 用户在线状态前缀
	UserOnlinePrefix = "user_online:"
	// 用户token前缀
	UserTokenPrefix = "user_token:"
)

// UserCache Redis实现的用户缓存
type UserCache struct {
	client *redis.Client
}

// NewUserCache 创建Redis用户缓存
func NewUserCache() UserCache {
	return UserCache{
		client: redis.GetClient(),
	}
}

// UserCache 用户缓存接口
type IUserCache interface {
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

// SetUser 缓存用户信息
func (c *UserCache) SetUser(ctx context.Context, user *entity.User) error {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%d", UserCachePrefix, user.ID)
	data, err := sonic.Marshal(user)
	if err != nil {
		logger.Errorf("序列化用户信息失败: %v", err)
		return err
	}
	err = c.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		logger.Errorf("缓存用户信息失败: %v", err)
		return err
	}
	logger.Debugf("用户信息已经缓存: userID=%d", user.ID)
	return nil
}

// GetUser 获取缓存的用户信息
func (c *UserCache) GetUser(ctx context.Context, userID uint64) (*entity.User, error) {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%d", UserCachePrefix, userID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var user entity.User
	if err := sonic.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser 删除用户缓存
func (c *UserCache) DeleteUser(ctx context.Context, userID uint64) error {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%d", UserCachePrefix, userID)
	return c.client.Del(ctx, key).Err()
}

// SetUserToken 缓存用户Token
func (c *UserCache) SetUserToken(ctx context.Context, userID uint64, tokenID string, token string, expiration time.Duration) error {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()

	// 存储token
	tokenKey := fmt.Sprintf("%s%d:%s", UserTokenPrefix, userID, tokenID)
	if err := c.client.Set(ctx, tokenKey, token, expiration).Err(); err != nil {
		logger.Errorf("缓存用户token失败: %v", err)
		return err
	}

	// 创建新的上下文用于第二个操作
	ctx2, cancel2 := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel2()

	// 存储用户的所有token ID列表
	listKey := fmt.Sprintf("%s%d:list", UserTokenPrefix, userID)
	if err := c.client.SAdd(ctx2, listKey, tokenID).Err(); err != nil {
		logger.Errorf("添加token ID到列表失败: %v", err)
		return err
	}

	return nil
}

// GetUserToken 获取用户Token
func (c *UserCache) GetUserToken(ctx context.Context, userID uint64, tokenID string) (string, error) {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()
	tokenKey := fmt.Sprintf("%s%d:%s", UserTokenPrefix, userID, tokenID)
	err := c.client.Get(ctx, tokenKey).Err()
	if err != nil {
		logger.Errorf("获取用户token失败: %v", err)
		return "", err
	}
	return tokenKey, nil

}

// RevokeUserToken 撤销指定的Token
func (c *UserCache) RevokeUserToken(ctx context.Context, userID uint64, tokenID string) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()

	tokenKey := fmt.Sprintf("%s%d:%s", UserTokenPrefix, userID, tokenID)
	if err := c.client.Del(ctx, tokenKey).Err(); err != nil {
		logger.Errorf("删除用户token失败: %v", err)
		return err
	}

	ctx2, cancel2 := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel2()

	listKey := fmt.Sprintf("%s%d:list", UserTokenPrefix, userID)
	if err := c.client.SRem(ctx2, listKey, tokenID).Err(); err != nil {
		logger.Errorf("从列表中移除token ID失败: %v", err)
		return err
	}
	return nil
}

// RevokeAllUserTokens 撤销用户的所有Token
func (c *UserCache) RevokeAllUserTokens(ctx context.Context, userID uint64) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()

	listKey := fmt.Sprintf("%s%d:list", UserTokenPrefix, userID)
	tokenIDs, err := c.client.SMembers(ctx, listKey).Result()
	if err != nil {
		logger.Errorf("获取用户token列表失败: %v", err)
		return err
	}
	for _, tokenID := range tokenIDs {
		delCtx, cancelDel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
		tokenKey := fmt.Sprintf("%s%d:%s", UserTokenPrefix, userID, tokenID)
		if err := c.client.Del(delCtx, tokenKey).Err(); err != nil {
			logger.Errorf("删除用户token失败: %v", err)
			return err
		}
		cancelDel()
	}
	listCtx, cancelList := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancelList()
	if err := c.client.Del(listCtx, listKey).Err(); err != nil {
		logger.Errorf("删除用户token列表失败: %v", err)
		return err
	}
	return nil
}

// SetUserOnline 设置用户在线状态
func (c *UserCache) SetUserOnline(ctx context.Context, userID uint64, deviceID string) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()
	key := fmt.Sprintf("%s%d", UserOnlinePrefix, userID)
	return c.client.HSet(ctx, key, deviceID, time.Now().Unix()).Err()
}

// SetUserOffline 设置用户离线状态
func (c *UserCache) SetUserOffline(ctx context.Context, userID uint64, deviceID string) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()
	key := fmt.Sprintf("%s%d", UserOnlinePrefix, userID)
	return c.client.HDel(ctx, key, deviceID).Err()
}

// IsUserOnline 检查用户是否在线
func (c *UserCache) IsUserOnline(ctx context.Context, userID uint64) (bool, error) {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, 500*time.Millisecond)
	defer cancel()
	key := fmt.Sprintf("%s%d", UserOnlinePrefix, userID)
	count, err := c.client.HLen(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
