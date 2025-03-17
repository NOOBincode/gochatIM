package messaging

import (
	"GochatIM/pkg/logger"
	"sync"

	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"
)

// KafkaProducer Kafka消息生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	mu       sync.Mutex
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	
	return &KafkaProducer{
		producer: producer,
	}, nil
}

// SendMessage 发送消息到Kafka
func (p *KafkaProducer) SendMessage(topic string, key string, value interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	jsonValue, err := sonic.Marshal(value)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return err
	}
	
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonValue),
	}
	
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		logger.Errorf("发送消息到Kafka失败: %v", err)
		return err
	}
	
	return nil
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	return p.producer.Close()
}