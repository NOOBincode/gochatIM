package event

import "time"

// CacheDeleteEvent 缓存删除事件
type CacheDeleteEvent struct {
	Type      string    `json:"type"`       // 缓存类型
	Key       string    `json:"key"`        // 缓存键
	Timestamp time.Time `json:"timestamp"`  // 事件时间戳
}