package entity

import (
	"time"
)

// MessageReceipt 消息接收状态模型
type MessageReceipt struct {
	ID        uint64     `gorm:"primaryKey;column:id" json:"id"`
	MessageID uint64     `gorm:"column:message_id;not null;index:idx_message_user,unique" json:"message_id"`
	UserID    uint64     `gorm:"column:user_id;not null;index:idx_message_user,unique;index:idx_user_id" json:"user_id"`
	Status    int8       `gorm:"column:status;type:tinyint(1);default:0" json:"status"` // 0-未读, 1-已读, 2-已删除
	ReadTime  *time.Time `gorm:"column:read_time" json:"read_time"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (MessageReceipt) TableName() string {
	return "message_receipts"
}
