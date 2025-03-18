package gateway

import (
	"net/http"
)

// MessageGateway 定义了消息网关的接口
type MessageGateway interface {
	// SendToUser 向指定用户发送消息
	SendToUser(userID uint64, message []byte) bool

	// IsUserOnline 检查用户是否在线
	IsUserOnline(userID uint64) bool

	// GetOnlineUserCount 获取在线用户数
	GetOnlineUserCount() int

	// GetUserConnectionCount 获取用户连接数
	GetUserConnectionCount(userID uint64) int

	// HandleWebsocket 处理WebSocket连接
	HandleWebsocket(w http.ResponseWriter, r *http.Request, userID uint64, deviceID string)
}
