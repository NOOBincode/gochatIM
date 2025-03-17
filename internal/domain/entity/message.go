package entity

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	MsgID        string    `gorm:"column:msg_id;size:50;not null;uniqueIndex:idx_msg_id" json:"msg_id"`
	SenderID     uint64    `gorm:"column:sender_id;not null;index:idx_sender_id" json:"sender_id"`
	ReceiverType int8      `gorm:"column:receiver_type;type:tinyint(1);not null;index:idx_receiver" json:"receiver_type"` // 0-个人, 1-群组
	ReceiverID   uint64    `gorm:"column:receiver_id;not null;index:idx_receiver" json:"receiver_id"`
	ContentType  int8      `gorm:"column:content_type;type:tinyint(1);not null;default:0" json:"content_type"` // 0-文本, 1-图片, 2-语音, 3-视频, 4-文件, 5-位置
	Content      string    `gorm:"column:content;type:text;not null" json:"content"`
	Extra        string    `gorm:"column:extra;type:text" json:"extra"`
	SendTime     time.Time `gorm:"column:send_time;not null;default:CURRENT_TIMESTAMP;index:idx_send_time" json:"send_time"`
	Status       int8      `gorm:"column:status;type:tinyint(1);default:0" json:"status"` // 0-未读, 1-已读, 2-已撤回, 3-已删除
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	
	// 关联
	Sender   User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receipts []MessageReceipt `gorm:"foreignKey:MessageID" json:"receipts,omitempty"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}