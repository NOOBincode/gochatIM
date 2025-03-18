package task

import (
	"context"
	"time"

	"GochatIM/internal/infrastructure/cache"
	"GochatIM/pkg/logger"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
)

const (
	// DeleteUserCache 删除用户缓存任务
	DeleteUserCache = "delete_user_cache"
)

// DeleteUserCachePayload 删除用户缓存任务载荷
type DeleteUserCachePayload struct {
	UserID uint64 `json:"user_id"`
}

// NewDeleteUserCacheTask 创建删除用户缓存任务
func NewDeleteUserCacheTask(userID uint64, delay time.Duration) (*asynq.Task, error) {
	payload, err := sonic.Marshal(DeleteUserCachePayload{UserID: userID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(DeleteUserCache, payload, asynq.ProcessIn(delay)), nil
}

// HandleDeleteUserCacheTask 处理删除用户缓存任务
func HandleDeleteUserCacheTask(ctx context.Context, t *asynq.Task, userCache cache.UserCache) error {
	var payload DeleteUserCachePayload
	if err := sonic.Unmarshal(t.Payload(), &payload); err != nil {
		logger.Errorf("解析删除用户缓存任务载荷失败: %v", err)
		return err
	}

	logger.Debugf("执行延迟删除用户缓存任务: userID=%d", payload.UserID)
	if err := userCache.DeleteUser(ctx, payload.UserID); err != nil {
		logger.Warnf("延迟删除用户缓存失败: %v", err)
		return err
	}

	return nil
}
