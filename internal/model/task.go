package model

import "time"

// Task 任务
type Task struct {
	BaseModel
	Name      string     `gorm:"size:128;not null" json:"name"`
	Type      string     `gorm:"size:64;not null" json:"type"` // batch_command / batch_check / custom
	Status    string     `gorm:"size:16;index" json:"status"` // pending / running / completed / failed / cancelled
	Payload   string     `gorm:"type:text" json:"payload"`     // JSON 任务参数
	Result    string     `gorm:"type:text" json:"result"`      // JSON 任务结果
	UserID    uint       `gorm:"index" json:"user_id"`
	Username  string     `gorm:"size:64" json:"username"`
	Total     int        `json:"total"`                          // 总数
	Done      int        `json:"done"`                           // 已完成
	StartedAt *time.Time `json:"started_at,omitempty"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}
