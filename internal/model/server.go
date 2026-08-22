package model

import "time"

// Server SSH 服务器模型
type Server struct {
	BaseModel
	Name           string     `gorm:"size:128;not null" json:"name"`
	Host           string     `gorm:"size:128;not null" json:"host"`
	Port           int        `gorm:"not null" json:"port"`
	Username       string     `gorm:"size:64;not null" json:"username"`
	Password       string     `gorm:"size:256" json:"-"`        // AES 加密存储
	PrivateKey     string     `gorm:"type:text" json:"-"`       // AES 加密存储
	Group          string     `gorm:"size:64" json:"group"`
	LastTestStatus string     `gorm:"size:16" json:"last_test_status"` // ok / fail
	LastTestTime   *time.Time `json:"last_test_time,omitempty"`
	LastTestMessage string    `gorm:"size:256" json:"last_test_message"`
}

func (Server) TableName() string {
	return "servers"
}
