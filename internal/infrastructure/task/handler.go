package task

import (
	"context"

	"GochatIM/internal/infrastructure/cache"
	"GochatIM/pkg/logger"

	"github.com/hibiken/asynq"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	userCache cache.UserCache
	// 其他依赖...
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(userCache cache.UserCache) *TaskHandler {
	return &TaskHandler{
		userCache: userCache,
	}
}

// RegisterHandlers 注册任务处理器
func (h *TaskHandler) RegisterHandlers(mux *asynq.ServeMux) {
	// 注册删除用户缓存任务处理器
	mux.HandleFunc(DeleteUserCache, func(ctx context.Context, t *asynq.Task) error {
		return HandleDeleteUserCacheTask(ctx, t, h.userCache)
	})

	// 注册其他任务处理器...

	logger.Info("任务处理器注册完成")
}
