package entity

import (
	"time"
)

// 消息状态常量
const (
	MessageStatusUnsent    = 0 // 未发送
	MessageStatusSent      = 1 // 已发送
	MessageStatusDelivered = 2 // 已送达
	MessageStatusRead      = 3 // 已读
	MessageStatusRecalled  = 4 // 已撤回
	MessageStatusDeleted   = 5 // 已删除
)

// 接收者类型常量
const (
	ReceiverTypeUser  = 0 // 个人
	ReceiverTypeGroup = 1 // 群组
)

// 内容类型常量
const (
	ContentTypeText     = 0 // 文本
	ContentTypeImage    = 1 // 图片
	ContentTypeVoice    = 2 // 语音
	ContentTypeVideo    = 3 // 视频
	ContentTypeFile     = 4 // 文件
	ContentTypeLocation = 5 // 位置
)

// Message 消息实体
type Message struct {
	ID             uint64    `json:"id"`              // 自增ID
	MsgID          string    `json:"msg_id"`          // 消息唯一标识
	ConversationID string    `json:"conversation_id"` // 会话ID
	SenderID       uint64    `json:"sender_id"`       // 发送者ID
	ReceiverType   int       `json:"receiver_type"`   // 接收者类型: 0-个人, 1-群组
	ReceiverID     uint64    `json:"receiver_id"`     // 接收者ID
	ContentType    int       `json:"content_type"`    // 内容类型
	Content        string    `json:"content"`         // 消息内容
	Extra          string    `json:"extra"`           // 附加信息(JSON格式)
	SendTime       time.Time `json:"send_time"`       // 发送时间
	Status         int       `json:"status"`          // 消息状态
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
	UpdatedAt      time.Time `json:"updated_at"`      // 更新时间
}
