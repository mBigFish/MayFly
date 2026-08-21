// Package target 定义目标（Target）实体及其仓储与服务。
package target

import "time"

// Target 是整个系统的核心实体，对应一个授权测试目标。
// 敏感字段（Cookies、Headers 中的认证信息等）需在存储前加密。
type Target struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	URL       string    `gorm:"size:2048;not null" json:"url"`
	Type      string    `gorm:"size:64" json:"type"`
	Protocol  string    `gorm:"size:64" json:"protocol"`
	Method    string    `gorm:"size:16" json:"method"`
	Headers   string    `gorm:"type:text" json:"headers"` // 加密存储，API 返回时由服务层控制是否暴露
	Cookies   string    `gorm:"type:text" json:"cookies"` // 加密存储，API 返回时由服务层控制是否暴露
	Timeout   int       `json:"timeout"`
	Proxy     string    `gorm:"size:2048" json:"proxy"`
	Encoding  string    `gorm:"size:64" json:"encoding"`
	Remark    string    `gorm:"type:text" json:"remark"`
	GroupID   uint      `gorm:"index" json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Target) TableName() string { return "targets" }

// TargetGroup 目标分组。
type TargetGroup struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Remark    string    `gorm:"type:text" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (TargetGroup) TableName() string { return "target_groups" }
