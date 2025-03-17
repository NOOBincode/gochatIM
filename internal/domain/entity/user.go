package entity

import (
	"time"
	
	"golang.org/x/crypto/bcrypt"
)

// User 用户实体
type User struct {
	ID        uint64
	Username  string
	Nickname  string
	Password  string // 存储的是加密后的密码哈希
	Email     string
	Phone     string
	Avatar    string
	Status    int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

// VerifyPassword 验证密码是否正确
func (u *User) VerifyPassword(plainPassword string) bool {
	// 使用 bcrypt 比较密码哈希
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))
	return err == nil
}

// SetPassword 设置密码（加密后存储）
func (u *User) SetPassword(plainPassword string) error {
	// 使用 bcrypt 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	u.Password = string(hashedPassword)
	return nil
}

// IsActive 检查用户是否处于活跃状态
func (u *User) IsActive() bool {
	return u.Status == 1
}

// Deactivate 停用用户账号
func (u *User) Deactivate() {
	u.Status = 0
	u.UpdatedAt = time.Now()
}

// Activate 激活用户账号
func (u *User) Activate() {
	u.Status = 1
	u.UpdatedAt = time.Now()
}

// HasEmail 检查用户是否设置了邮箱
func (u *User) HasEmail() bool {
	return u.Email != ""
}

// HasPhone 检查用户是否设置了手机号
func (u *User) HasPhone() bool {
	return u.Phone != ""
}

// SanitizedUser 返回不包含敏感信息的用户数据
func (u *User) SanitizedUser() map[string]interface{} {
	return map[string]interface{}{
		"id":        u.ID,
		"username":  u.Username,
		"nickname":  u.Nickname,
		"email":     u.Email,
		"phone":     u.Phone,
		"avatar":    u.Avatar,
		"status":    u.Status,
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
	}
}