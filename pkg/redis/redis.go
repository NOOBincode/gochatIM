package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

// Setup 初始化Redis连接
func Setup() error {
	Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.pool_size"),
		MinIdleConns: viper.GetInt("redis.min_idle_conns"),
		DialTimeout:  time.Duration(viper.GetInt("redis.dial_timeout")) * time.Second,
		ReadTimeout:  time.Duration(viper.GetInt("redis.read_timeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("redis.write_timeout")) * time.Second,
	})

	_, err := Client.Ping(Ctx).Result()
	return err
}

// Close 关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}