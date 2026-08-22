package model

// Listener 反向 Shell 监听器
type Listener struct {
	BaseModel
	Name        string `gorm:"size:128;not null" json:"name"`
	Host        string `gorm:"size:64;not null" json:"host"`
	Port        int    `gorm:"not null" json:"port"`
	Protocol    string `gorm:"size:16" json:"protocol"` // tcp / udp
	Status      string `gorm:"size:16" json:"status"`   // running / stopped
	Connections int    `json:"connections"`
	UserID      uint   `gorm:"index" json:"user_id"`
}

func (Listener) TableName() string {
	return "listeners"
}
