package model

import "time"

// Session 操作会话
type Session struct {
	BaseModel
	TargetID   uint      `gorm:"index;not null" json:"target_id"`
	TargetName string    `gorm:"size:128" json:"target_name"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Username   string    `gorm:"size:64" json:"username"`
	Type       string    `gorm:"size:32;not null" json:"type"` // command / file / terminal / db
	Status     string    `gorm:"size:16" json:"status"`       // active / closed
	LastActive *time.Time `json:"last_active,omitempty"`
}

func (Session) TableName() string {
	return "sessions"
}
