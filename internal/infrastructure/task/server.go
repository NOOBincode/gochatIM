package task

import (
	"GochatIM/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

// Server 任务服务器
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer 创建任务服务器
func NewServer() *Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     viper.GetString("redis.host") + ":" + viper.GetString("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10, // 并发处理任务的数量
			Logger:      logger.NewLogger().Sugar(),
		},
	)

	mux := asynq.NewServeMux()

	return &Server{
		server: server,
		mux:    mux,
	}
}

// RegisterHandler 注册任务处理器
func (s *Server) RegisterHandler(handler *TaskHandler) {
	handler.RegisterHandlers(s.mux)
}

// Start 启动服务器
func (s *Server) Start() error {
	logger.Info("任务服务器启动")
	return s.server.Run(s.mux)
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	logger.Info("任务服务器关闭")
	s.server.Shutdown()
}
