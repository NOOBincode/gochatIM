package messaging

import (
	"GochatIM/internal/domain/event"
	"GochatIM/internal/infrastructure/cache"
	"GochatIM/pkg/logger"
	"context"
	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"
	"strconv"
)

// CacheDeleteConsumer 缓存删除事件消费者
type CacheDeleteConsumer struct {
	userCache cache.UserCache
}

// NewCacheDeleteConsumer 创建缓存删除事件消费者
func NewCacheDeleteConsumer(userCache cache.UserCache) *CacheDeleteConsumer {
	return &CacheDeleteConsumer{
		userCache: userCache,
	}
}

// Setup 实现 sarama.ConsumerGroupHandler 接口
func (c *CacheDeleteConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 实现 sarama.ConsumerGroupHandler 接口
func (c *CacheDeleteConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消息
func (c *CacheDeleteConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var delayEvent event.CacheDeleteEvent
		if err := sonic.Unmarshal(msg.Value, &delayEvent); err != nil {
			logger.Errorf("解析缓存删除事件失败: %v", err)
			continue
		}

		// 处理不同类型的缓存删除
		switch delayEvent.Type {
		case "user_cache":
			userID, err := strconv.ParseUint(delayEvent.Key, 10, 64)
			if err != nil {
				logger.Errorf("解析用户ID失败: %v", err)
				continue
			}
			if err := c.userCache.DeleteUser(context.Background(), userID); err != nil {
				logger.Warnf("消费者删除用户缓存失败: %v", err)
			}
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
