package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"GochatIM/pkg/logger"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var (
	// RedisClient 全局Redis客户端实例（保留向后兼容性）
	RedisClient *redis.Client
	// Ctx 全局上下文（保留向后兼容性）
	Ctx = context.Background()
	// 单例锁
	once sync.Once
	// 默认客户端实例
	defaultClient *Client
)

// Client Redis客户端封装
type Client struct {
	*redis.Client
	config *Config
}

// Config Redis配置
type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:         viper.GetString("redis.host"),
		Port:         viper.GetInt("redis.port"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.pool_size"),
		MinIdleConns: viper.GetInt("redis.min_idle_conns"),
		DialTimeout:  time.Duration(viper.GetInt("redis.dial_timeout")) * time.Second,
		ReadTimeout:  time.Duration(viper.GetInt("redis.read_timeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("redis.write_timeout")) * time.Second,
	}
}

// NewClient 创建新的Redis客户端
func NewClient() (*Client, error) {
	var config *Config
	if config == nil {
		config = DefaultConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Errorf("Redis连接失败: %v", err)
		return nil, err
	}

	logger.Info("Redis连接成功")
	return &Client{
		Client: client,
		config: config,
	}, nil
}

// Setup 初始化Redis连接（保留向后兼容性）
func Setup() error {
	var err error
	once.Do(func() {
		var client *Client
		client, err = NewClient()
		if err != nil {
			return
		}

		// 设置全局变量，保持向后兼容
		RedisClient = client.Client
		defaultClient = client
	})
	return err
}

// Close 关闭Redis连接（保留向后兼容性）
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// GetClient 获取默认客户端实例
func GetClient() *Client {
	if defaultClient == nil {
		if err := Setup(); err != nil {
			logger.Fatalf("初始化Redis客户端失败: %v", err)
		}
	}
	return defaultClient
}

// CreateTimeoutContext 创建带超时的上下文
func (c *Client) CreateTimeoutContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}

// WithTimeout 执行带超时的Redis操作
func (c *Client) WithTimeout(timeout time.Duration, fn func(ctx context.Context) error) error {
	ctx, cancel := c.CreateTimeoutContext(nil, timeout)
	defer cancel()

	return fn(ctx)
}
