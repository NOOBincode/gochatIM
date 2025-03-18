package cache

import (
	"GochatIM/internal/domain/entity"
	"GochatIM/pkg/logger"
	"GochatIM/pkg/redis"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	redislib "github.com/go-redis/redis/v8" // 直接导入redis库
)

// MessageCache 消息缓存接口
type IMessageCache interface {
	// SetMessage 缓存单条消息
	SetMessage(ctx context.Context, message *entity.Message) error

	// GetMessage 获取缓存的消息
	GetMessage(ctx context.Context, msgID string) (*entity.Message, error)

	// AddRecentMessage 添加消息到最近消息列表
	AddRecentMessage(ctx context.Context, receiverType int8, receiverID uint64, message *entity.Message) error

	// GetRecentMessages 获取最近消息列表
	GetRecentMessages(ctx context.Context, receiverType int8, receiverID uint64, offset, limit int64) ([]*entity.Message, error)

	// SetMessageReadStatus 设置消息已读状态
	SetMessageReadStatus(ctx context.Context, msgID string, userID uint64) error

	// GetMessageReadUsers 获取已读消息的用户列表
	GetMessageReadUsers(ctx context.Context, msgID string) ([]uint64, error)

	// DeleteMessage 删除缓存的消息
	DeleteMessage(ctx context.Context, msgID string) error

	// ClearRecentMessages 清空最近消息列表
	ClearRecentMessages(ctx context.Context, receiverType int8, receiverID uint64) error
}

const (
	// 消息缓存前缀
	MessageCachePrefix = "message:"
	// 最近消息列表前缀
	RecentMessagePrefix = "recent_messages:"
	// 默认缓存操作超时时间（毫秒）
	defaultCacheTimeout = 500
)

// RedisMessageCache Redis实现的消息缓存
type MessageCache struct {
	client *redis.Client
}

// NewRedisMessageCache 创建Redis消息缓存
func NewMessageCache() *MessageCache {
	return &MessageCache{
		client: redis.GetClient(),
	}
}

// SetMessage 缓存单条消息
func (c *MessageCache) SetMessage(ctx context.Context, message *entity.Message) error {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%s", MessageCachePrefix, message.MsgID)
	data, err := sonic.Marshal(message)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return err
	}

	// 设置缓存时间，默认24小时
	cacheTime := 24 * time.Hour

	_, err = c.client.Set(ctx, key, data, cacheTime).Result()
	if err != nil {
		logger.Errorf("缓存消息失败: %v", err)
		return err
	}
	logger.Debugf("缓存消息成功: %v", message)
	return nil
}

// GetMessage 获取缓存的消息
func (c *MessageCache) GetMessage(ctx context.Context, msgID string) (*entity.Message, error) {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%s", MessageCachePrefix, msgID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var message entity.Message
	if err := sonic.Unmarshal(data, &message); err != nil {
		logger.Errorf("解析消息失败: %v", err)
		return nil, err
	}
	logger.Debugf("获取消息成功: %v", message)
	return &message, nil
}

// AddRecentMessage 添加消息到最近消息列表
func (c *MessageCache) AddRecentMessage(ctx context.Context, receiverType int8, receiverID uint64, message *entity.Message) error {
	// 创建带超时的上下文
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%d:%d", RecentMessagePrefix, receiverType, receiverID)
	data, err := sonic.Marshal(message)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return err
	}

	// 使用消息发送时间作为分数，便于按时间排序
	score := float64(message.SendTime.Unix())

	// 添加到有序集合
	if err := c.client.ZAdd(ctx, key, &redislib.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		logger.Errorf("添加消息到最近消息列表失败: %v", err)
		return err
	}

	// 保留最近N条消息
	historyCount := int64(100) // 默认保留100条

	// 创建新的超时上下文，因为这是另一个操作
	ctx2, cancel2 := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel2()

	count, err := c.client.ZCard(ctx2, key).Result()
	if err != nil {
		logger.Errorf("获取消息列表数量失败: %v", err)
		return err
	}

	if count > historyCount {
		// 创建新的超时上下文，因为这是另一个操作
		ctx3, cancel3 := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
		defer cancel3()

		// 删除多余的旧消息
		if err := c.client.ZRemRangeByRank(ctx3, key, 0, count-historyCount-1).Err(); err != nil {
			logger.Errorf("删除旧消息失败: %v", err)
			return err
		}
	}
	logger.Debugf("添加消息到最近消息列表成功: %v", message)
	return nil
}

// GetRecentMessages 获取最近消息列表
func (c *MessageCache) GetRecentMessages(ctx context.Context, receiverType int8, receiverID uint64, offset, limit int64) ([]*entity.Message, error) {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%d:%d", RecentMessagePrefix, receiverType, receiverID)

	// 按时间倒序获取消息
	result, err := c.client.ZRevRange(ctx, key, offset, offset+limit-1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]*entity.Message, 0, len(result))
	for _, data := range result {
		var message entity.Message
		if err := sonic.Unmarshal([]byte(data), &message); err != nil {
			continue
		}
		messages = append(messages, &message)
	}
	logger.Debugf("获取最近消息列表成功: %v", messages)
	return messages, nil
}

// SetMessageReadStatus 设置消息已读状态
func (c *MessageCache) SetMessageReadStatus(ctx context.Context, msgID string, userID uint64) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%s:read", MessageCachePrefix, msgID)
	_, err := c.client.SAdd(ctx, key, strconv.FormatUint(userID, 10)).Result()
	if err != nil {
		logger.Errorf("设置消息已读状态失败: %v", err)
		return err
	}
	logger.Debugf("设置消息已读状态成功: %v", key)
	return nil
}

// GetMessageReadUsers 获取已读消息的用户列表
func (c *MessageCache) GetMessageReadUsers(ctx context.Context, msgID string) ([]uint64, error) {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("%s%s:read", MessageCachePrefix, msgID)
	result, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		logger.Errorf("获取已读消息的用户列表失败: %v", err)
		return nil, err
	}

	userIDs := make([]uint64, 0, len(result))
	for _, id := range result {
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}
	logger.Debugf("获取已读消息的用户列表成功: %v", userIDs)
	return userIDs, nil
}

// ClearRecentMessages implements cache.MessageCache.
func (c *MessageCache) ClearRecentMessages(ctx context.Context, receiverType int8, receiverID uint64) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()
	key := fmt.Sprintf("%s%d:%d", RecentMessagePrefix, receiverType, receiverID)
	// 删除有序集合
	if err := c.client.Del(ctx, key).Err(); err != nil {
		logger.Errorf("删除最近消息列表失败: %v", err)
		return err
	}
	logger.Debugf("删除最近消息列表成功: %v", key)
	return nil
}

// DeleteMessage implements cache.MessageCache.
func (c *MessageCache) DeleteMessage(ctx context.Context, msgID string) error {
	ctx, cancel := c.client.CreateTimeoutContext(ctx, defaultCacheTimeout*time.Millisecond)
	defer cancel()
	key := fmt.Sprintf("%s%s", MessageCachePrefix, msgID)
	// 删除消息
	if err := c.client.Del(ctx, key).Err(); err != nil {
		logger.Errorf("删除消息失败: %v", err)
		return err
	}
	logger.Debugf("删除消息成功: %v", key)
	return nil
}
