package messaging

import (
	"GochatIM/internal/domain/event"
	"GochatIM/internal/infrastructure/cache"
	"GochatIM/pkg/logger"
	"context"
	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"
	"sync"
)

// UserEventConsumer 用户事件消费者
type UserEventConsumer struct {
	consumerGroup sarama.ConsumerGroup
	userCache     cache.UserCache
	topics        []string
	ready         chan bool
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

// NewUserEventConsumer 创建用户事件消费者
func NewUserEventConsumer(brokers []string, groupID string, topics []string, userCache cache.UserCache) (*UserEventConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return &UserEventConsumer{
		consumerGroup: consumerGroup,
		userCache:     userCache,
		topics:        topics,
		ready:         make(chan bool),
		ctx:           ctx,
		cancelFunc:    cancel,
	}, nil
}

type userConsumerHandler struct {
	userCache cache.UserCache
	ready     chan bool
}

// Setup 初始化消费者
func (h *userConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	select {
	case h.ready <- true:
	default:
	}
	return nil
}

func (h *userConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *userConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var userEvent event.UserEvent
		if err := sonic.Unmarshal(message.Value, &userEvent); err != nil {
			logger.Errorf("解析用户事件失败: %v", err)
			continue
		}

		logger.Infof("收到用户事件: type=%s, userID=%d", userEvent.EventType, userEvent.UserID)

		switch userEvent.EventType {
		case event.UserCreated, event.UserUpdated:
			// 这里可以直接从数据库加载最新数据更新缓存
			// 或者事件中包含完整的用户数据直接更新缓存
			// 简化处理：这里我们只删除缓存，让下次查询时重新加载
			if err := h.userCache.DeleteUser(context.Background(), userEvent.UserID); err != nil {
				logger.Warnf("删除用户缓存失败: %v", err)
			}
		case event.UserDeleted:
			// 删除用户缓存
			if err := h.userCache.DeleteUser(context.Background(), userEvent.UserID); err != nil {
				logger.Warnf("删除用户缓存失败: %v", err)
			}
			// 撤销所有token
			if err := h.userCache.RevokeAllUserTokens(context.Background(), userEvent.UserID); err != nil {
				logger.Warnf("撤销用户所有token失败: %v", err)
			}
		default:
			logger.Warnf("未知的用户事件类型: %s", userEvent.EventType)
		}

		session.MarkMessage(message, "")
	}
	return nil
}

// Start 启动消费者
func (c *UserEventConsumer) Start() error {
	handler := &userConsumerHandler{
		userCache: c.userCache,
		ready:     c.ready,
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			if err := c.consumerGroup.Consume(c.ctx, c.topics, handler); err != nil {
				logger.Errorf("消费用户事件失败: %v", err)
			}

			if c.ctx.Err() != nil {
				return
			}

			logger.Infof("用户事件消费者重新连接中...")
		}
	}()
	<-c.ready
	logger.Infof("用户事件消费者已就绪")
	return nil
}

// Stop 停止消费者
func (c *UserEventConsumer) Stop() error {
	c.cancelFunc()
	return c.consumerGroup.Close()
}

// Ready 返回就绪通道
func (c *UserEventConsumer) Ready() <-chan bool {
	return c.ready
}
