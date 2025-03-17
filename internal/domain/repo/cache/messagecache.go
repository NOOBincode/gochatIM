package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	redislib "github.com/go-redis/redis/v8" // 直接导入redis库
	"GochatIM/internal/domain/entity"
	"GochatIM/pkg/redis"
	"github.com/spf13/viper"
)

const (
	// 消息缓存前缀
	MessageCachePrefix = "message:"
	// 最近消息列表前缀
	RecentMessagePrefix = "recent_messages:"
	// 默认缓存操作超时时间（毫秒）
	defaultCacheTimeout = 500
)

// 创建带超时的上下文
func createTimeoutContextMessage() (context.Context, context.CancelFunc) {
	timeout := time.Duration(viper.GetInt("redis.operation_timeout")) * time.Millisecond
	if timeout == 0 {
		timeout = defaultCacheTimeout * time.Millisecond
	}
	return context.WithTimeout(redis.Ctx, timeout)
}

// SetMessage 缓存单条消息
func SetMessage(message *entity.Message) error {
	ctx, cancel := createTimeoutContextMessage()
	defer cancel()
	
	key := fmt.Sprintf("%s%s", MessageCachePrefix, message.MsgID)
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	cacheTime := time.Duration(viper.GetInt("message.cache_time")) * time.Second
	return redis.Client.Set(ctx, key, data, cacheTime).Err()
}

// GetMessage 获取缓存的消息
func GetMessage(msgID string) (*entity.Message, error) {
	ctx, cancel := createTimeoutContextMessage()
	defer cancel()
	
	key := fmt.Sprintf("%s%s", MessageCachePrefix, msgID)
	data, err := redis.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	
	var message entity.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}
	
	return &message, nil
}

// AddRecentMessage 添加消息到最近消息列表
func AddRecentMessage(receiverType int8, receiverID uint64, message *entity.Message) error {
	ctx, cancel := createTimeoutContextMessage()
	defer cancel()
	
	key := fmt.Sprintf("%s%d:%d", RecentMessagePrefix, receiverType, receiverID)
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	// 使用消息发送时间作为分数，便于按时间排序
	score := float64(message.SendTime.Unix())
	
	// 添加到有序集合
	if err := redis.Client.ZAdd(ctx, key, &redislib.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return err
	}
	
	// 保留最近N条消息
	historyCount := viper.GetInt64("message.history_count")
	
	// 创建新的超时上下文，因为这是另一个操作
	ctx2, cancel2 := createTimeoutContextMessage()
	defer cancel2()
	
	count, err := redis.Client.ZCard(ctx2, key).Result()
	if err != nil {
		return err
	}
	
	if count > historyCount {
		// 创建新的超时上下文，因为这是另一个操作
		ctx3, cancel3 := createTimeoutContextMessage()
		defer cancel3()
		
		// 删除多余的旧消息
		return redis.Client.ZRemRangeByRank(ctx3, key, 0, count-historyCount-1).Err()
	}
	
	return nil
}

// GetRecentMessages 获取最近消息列表
func GetRecentMessages(receiverType int8, receiverID uint64, offset, limit int64) ([]*entity.Message, error) {
	ctx, cancel := createTimeoutContextMessage()
	defer cancel()
	
	key := fmt.Sprintf("%s%d:%d", RecentMessagePrefix, receiverType, receiverID)
	
	// 按时间倒序获取消息
	result, err := redis.Client.ZRevRange(ctx, key, offset, offset+limit-1).Result()
	if err != nil {
		return nil, err
	}
	
	messages := make([]*entity.Message, 0, len(result))
	for _, data := range result {
		var message entity.Message
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			continue
		}
		messages = append(messages, &message)
	}
	
	return messages, nil
}

// SetMessageReadStatus 设置消息已读状态
func SetMessageReadStatus(msgID string, userID uint64) error {
	ctx, cancel := createTimeoutContextMessage()
	defer cancel()
	
	key := fmt.Sprintf("%s%s:read", MessageCachePrefix, msgID)
	return redis.Client.SAdd(ctx, key, strconv.FormatUint(userID, 10)).Err()
}

// GetMessageReadUsers 获取已读消息的用户列表
func GetMessageReadUsers(msgID string) ([]uint64, error) {
	ctx, cancel := createTimeoutContextMessage()
	defer cancel()
	
	key := fmt.Sprintf("%s%s:read", MessageCachePrefix, msgID)
	result, err := redis.Client.SMembers(ctx, key).Result()
	if err != nil {
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
	
	return userIDs, nil
}