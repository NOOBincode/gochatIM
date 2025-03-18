package messaging

import (
	"context"
	"sync"

	"GochatIM/internal/domain/gateway"
	"GochatIM/internal/domain/model"
	"GochatIM/pkg/logger"

	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"
)

// MessageHandler 消息处理接口
type MessageHandler interface {
	// ProcessIncomingMessage 处理接收到的消息
	ProcessIncomingMessage(ctx context.Context, message *model.Message) error
}

// GatewayMessageConsumer 消息网关消费者
type GatewayMessageConsumer struct {
	consumer       sarama.Consumer
	gateway        gateway.MessageGateway
	topics         []string
	ready          chan bool
	ctx            context.Context
	cancelFunc     context.CancelFunc
	messageHandler MessageHandler
}

// NewGatewayMessageConsumer 创建消息网关消费者
func NewGatewayMessageConsumer(brokers []string, topics []string, gateway gateway.MessageGateway, messageHandler MessageHandler) (*GatewayMessageConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &GatewayMessageConsumer{
		consumer:       consumer,
		gateway:        gateway,
		topics:         topics,
		ready:          make(chan bool),
		ctx:            ctx,
		cancelFunc:     cancel,
		messageHandler: messageHandler,
	}, nil
}

// Start 启动消费者
func (c *GatewayMessageConsumer) Start() error {
	var wg sync.WaitGroup

	// 为每个主题创建一个消费者
	for _, topic := range c.topics {
		partitions, err := c.consumer.Partitions(topic)
		if err != nil {
			logger.Errorf("获取主题分区失败: %v", err)
			return err
		}

		// 为每个分区创建一个消费者
		for _, partition := range partitions {
			pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
			if err != nil {
				logger.Errorf("创建分区消费者失败: %v", err)
				return err
			}

			wg.Add(1)
			go func(pc sarama.PartitionConsumer) {
				defer wg.Done()
				defer pc.Close()

				// 消费消息
				for {
					select {
					case msg := <-pc.Messages():
						c.handleMessage(msg)
					case err := <-pc.Errors():
						logger.Errorf("消费消息错误: %v", err)
					case <-c.ctx.Done():
						return
					}
				}
			}(pc)
		}
	}

	c.ready <- true
	wg.Wait()
	return nil
}

// Stop 停止消费者
func (c *GatewayMessageConsumer) Stop() error {
	c.cancelFunc()
	return c.consumer.Close()
}

// Ready 返回就绪通道
func (c *GatewayMessageConsumer) Ready() <-chan bool {
	return c.ready
}

// 处理消息
func (c *GatewayMessageConsumer) handleMessage(msg *sarama.ConsumerMessage) {
	var wsMessage model.Message
	if err := sonic.Unmarshal(msg.Value, &wsMessage); err != nil {
		logger.Errorf("解析WebSocket消息失败: %v", err)
		return
	}

	logger.Debugf("收到WebSocket消息: op=%d, sender=%d, receiver=%d",
		wsMessage.Operation, wsMessage.SenderID, wsMessage.ReceiverID)

	// 处理消息
	if err := c.messageHandler.ProcessIncomingMessage(context.Background(), &wsMessage); err != nil {
		logger.Errorf("处理WebSocket消息失败: %v", err)
		return
	}

	// 如果接收者在线，直接投递消息
	if wsMessage.ReceiverID > 0 && c.gateway.IsUserOnline(wsMessage.ReceiverID) {
		data, err := sonic.Marshal(wsMessage)
		if err != nil {
			logger.Errorf("序列化WebSocket消息失败: %v", err)
			return
		}

		c.gateway.SendToUser(wsMessage.ReceiverID, data)
		logger.Debugf("消息已投递给接收者: userID=%d", wsMessage.ReceiverID)
	}
}
