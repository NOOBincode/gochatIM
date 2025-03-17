package cache

import (
	"context"
	"fmt"
	"time"
	
	"GochatIM/internal/domain/entity"
	"GochatIM/internal/domain/repo/cache"
	"GochatIM/pkg/logger"
	"GochatIM/pkg/redis"
	
	"github.com/bytedance/sonic"
	"github.com/spf13/viper"
)

const (
	// 默认缓存超时时间
	defaultCacheTimeout = 500 // 毫秒
)

// RedisUserCache Redis实现的用户缓存
type RedisUserCache struct{}

// NewRedisUserCache 创建Redis用户缓存
func NewRedisUserCache() cache.UserCache {
	return &RedisUserCache{}
}

func createTimeoutContextUser() (context.Context, context.CancelFunc) {
	timeout := time.Duration(viper.GetInt("redis.operation_timeout")) * time.Millisecond
	if timeout == 0 {
		timeout = defaultCacheTimeout * time.Millisecond
	}
	return context.WithTimeout(context.Background(), timeout)
}

// SetUser 缓存用户信息
func (c *RedisUserCache) SetUser(ctx context.Context, user *entity.User) error {
	key := fmt.Sprintf("%s%d", cache.UserCachePrefix, user.ID)
	data, err := sonic.Marshal(user)
	if err != nil {
		logger.Errorf("序列化用户信息失败: %v", err)
		return err
	}
	err = redis.Client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		logger.Errorf("缓存用户信息失败: %v", err)
		return err
	}
	logger.Debugf("用户信息已经缓存: userID=%d", user.ID)
	return nil
}

// GetUser 获取缓存的用户信息
func (c *RedisUserCache) GetUser(ctx context.Context, userID uint64) (*entity.User, error) {
	key := fmt.Sprintf("%s%d", cache.UserCachePrefix, userID)
	data, err := redis.Client.Get(ctx, key).Bytes()
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
func (c *RedisUserCache) DeleteUser(ctx context.Context, userID uint64) error {
	key := fmt.Sprintf("%s%d", cache.UserCachePrefix, userID)
	return redis.Client.Del(ctx, key).Err()
}

// SetUserToken 缓存用户Token
func (c *RedisUserCache) SetUserToken(ctx context.Context, userID uint64, tokenID string, token string, expiration time.Duration) error {
	// 存储token
	tokenKey := fmt.Sprintf("%s%d:%s", cache.UserTokenPrefix, userID, tokenID)
	if err := redis.Client.Set(ctx, tokenKey, token, expiration).Err(); err != nil {
		logger.Errorf("缓存用户token失败: %v", err)
		return err
	}
	
	// 存储用户的所有token ID列表
	listKey := fmt.Sprintf("%s%d:list", cache.UserTokenPrefix, userID)
	if err := redis.Client.SAdd(ctx, listKey, tokenID).Err(); err != nil {
		logger.Errorf("添加token ID到列表失败: %v", err)
		return err
	}
	
	return nil
}

// GetUserToken 获取用户Token
func (c *RedisUserCache) GetUserToken(ctx context.Context, userID uint64, tokenID string) (string, error) {
	tokenKey := fmt.Sprintf("%s%d:%s", cache.UserTokenPrefix, userID, tokenID)
	result, err := redis.Client.Get(ctx, tokenKey).Result()
	if err != nil {
		logger.Errorf("获取用户特定token失败: %v", err)
		return "", err
	}
	return result, nil
}

// RevokeUserToken 撤销指定的Token
func (c *RedisUserCache) RevokeUserToken(ctx context.Context, userID uint64, tokenID string) error {
	// 删除特定token
	tokenKey := fmt.Sprintf("%s%d:%s", cache.UserTokenPrefix, userID, tokenID)
	if err := redis.Client.Del(ctx, tokenKey).Err(); err != nil {
		logger.Errorf("撤销用户token失败: %v", err)
		return err
	}
	
	// 从列表中移除
	listKey := fmt.Sprintf("%s%d:list", cache.UserTokenPrefix, userID)
	if err := redis.Client.SRem(ctx, listKey, tokenID).Err(); err != nil {
		logger.Errorf("从列表移除token ID失败: %v", err)
		return err
	}
	
	return nil
}

// RevokeAllUserTokens 撤销用户的所有Token
func (c *RedisUserCache) RevokeAllUserTokens(ctx context.Context, userID uint64) error {
	// 获取用户的所有token ID
	listKey := fmt.Sprintf("%s%d:list", cache.UserTokenPrefix, userID)
	tokenIDs, err := redis.Client.SMembers(ctx, listKey).Result()
	if err != nil {
		logger.Errorf("获取用户token列表失败: %v", err)
		return err
	}
	
	// 删除所有token
	for _, tokenID := range tokenIDs {
		tokenKey := fmt.Sprintf("%s%d:%s", cache.UserTokenPrefix, userID, tokenID)
		if err := redis.Client.Del(ctx, tokenKey).Err(); err != nil {
			logger.Errorf("删除用户token失败: %v", err)
			// 继续删除其他token
		}
	}
	
	// 删除列表
	if err := redis.Client.Del(ctx, listKey).Err(); err != nil {
		logger.Errorf("删除用户token列表失败: %v", err)
		return err
	}
	
	return nil
}

// SetUserOnline 设置用户在线状态
func (c *RedisUserCache) SetUserOnline(ctx context.Context, userID uint64, deviceID string) error {
	key := fmt.Sprintf("%s%d", cache.UserOnlinePrefix, userID)
	// 使用Hash存储多设备在线状态
	return redis.Client.HSet(ctx, key, deviceID, time.Now().Unix()).Err()
}

// SetUserOffline 设置用户离线状态
func (c *RedisUserCache) SetUserOffline(ctx context.Context, userID uint64, deviceID string) error {
	key := fmt.Sprintf("%s%d", cache.UserOnlinePrefix, userID)
	return redis.Client.HDel(ctx, key, deviceID).Err()
}

// IsUserOnline 检查用户是否在线
func (c *RedisUserCache) IsUserOnline(ctx context.Context, userID uint64) (bool, error) {
	key := fmt.Sprintf("%s%d", cache.UserOnlinePrefix, userID)
	count, err := redis.Client.HLen(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}