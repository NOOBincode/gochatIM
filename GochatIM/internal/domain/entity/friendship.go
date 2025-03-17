package entity

import (
	"time"
)

// Friendship 好友关系模型
type Friendship struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;index:idx_user_friend,unique" json:"user_id"`
	FriendID  uint64    `gorm:"column:friend_id;not null;index:idx_user_friend,unique;index:idx_friend_id" json:"friend_id"`
	Remark    string    `gorm:"size:50" json:"remark"`
	Status    int8      `gorm:"type:tinyint(1);default:1" json:"status"` // 0-待确认, 1-已确认, 2-已拒绝, 3-已拉黑
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	
	// 关联
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Friend User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}

// TableName 指定表名
func (Friendship) TableName() string {
	return "friendships"
}