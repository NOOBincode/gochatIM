package task

import (
	"context"
	"time"

	"GochatIM/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

// Client 任务客户端
type Client struct {
	client *asynq.Client
}

// NewClient 创建任务客户端
func NewClient() *Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     viper.GetString("redis.host") + ":" + viper.GetString("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}

	client := asynq.NewClient(redisOpt)
	return &Client{
		client: client,
	}
}

// Enqueue 提交任务
func (c *Client) Enqueue(ctx context.Context, taskType string, payload interface{}, delay time.Duration) (*asynq.TaskInfo, error) {
	var task *asynq.Task
	var err error

	switch taskType {
	case DeleteUserCache:
		if p, ok := payload.(DeleteUserCachePayload); ok {
			task, err = NewDeleteUserCacheTask(p.UserID, delay)
		} else {
			logger.Errorf("无效的任务载荷类型: %T", payload)
			return nil, ErrInvalidPayload
		}
	// 可以添加其他任务类型的处理
	default:
		logger.Errorf("未知的任务类型: %s", taskType)
		return nil, ErrUnknownTaskType
	}

	if err != nil {
		return nil, err
	}

	return c.client.EnqueueContext(ctx, task)
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.client.Close()
}

// 错误定义
var (
	ErrInvalidPayload  = asynq.ErrTaskNotFound
	ErrUnknownTaskType = asynq.ErrTaskNotFound
)
