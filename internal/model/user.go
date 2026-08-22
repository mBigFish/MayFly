package model

import "time"

// User 用户模型
type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password string `gorm:"size:128;not null" json:"-"` // bcrypt 哈希
	Nickname string `gorm:"size:64" json:"nickname"`
	Email    string `gorm:"size:128" json:"email"`
	Role     string `gorm:"size:32;not null" json:"role"` // admin / operator / auditor
	Status   int    `gorm:"not null" json:"status"`       // 1=启用 0=禁用
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

func (User) TableName() string {
	return "users"
}
