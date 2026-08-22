package model

// CmdHistory 命令执行历史
type CmdHistory struct {
	BaseModel
	TargetID uint   `gorm:"index;not null" json:"target_id"`
	UserID   uint   `gorm:"index" json:"user_id"`
	Username string `gorm:"size:64" json:"username"`
	Command  string `gorm:"type:text;not null" json:"command"`
	Output   string `gorm:"type:text" json:"output"`
	Error    string `gorm:"type:text" json:"error,omitempty"`
}

func (CmdHistory) TableName() string {
	return "cmd_history"
}
