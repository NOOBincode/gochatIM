package event

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// 用户事件类型
const (
	UserCreated = "user.created"
	UserUpdated = "user.updated"
	UserDeleted = "user.deleted"
)

// UserEvent 用户事件
type UserEvent struct {
	EventType  string    `json:"event_type"`
	UserID     uint64    `json:"user_id"`
	Username   string    `json:"username"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	EventID    string    `json:"event_id"`
	SequenceID int64     `json:"sequence_id,omitempty"`
}

// NewUserEvent 创建用户事件
func NewUserEvent(eventType string, userID uint64, username string) *UserEvent {
	return &UserEvent{
		EventType: eventType,
		UserID:    userID,
		Username:  username,
		Timestamp: time.Now(),
		EventID:   generateEventID(), // 实现一个生成唯一事件ID的函数
	}
}

// 生成唯一事件ID
func generateEventID() string {
	id := uuid.New()
	return time.Now().Format("20060102150405") + "-" + id.String()
}

// 生成随机字符串
func randomString(length int) string {
	bytes := make([]byte,length/2)
	//使用加密安全的随机数生成器
	if _,err := rand.Read(bytes);err != nil{
		return uuid.New().String()[:length]
	}
	return hex.EncodeToString(bytes)
}