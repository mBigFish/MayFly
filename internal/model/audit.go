package model

// AuditLog 审计日志
type AuditLog struct {
	BaseModel
	UserID   uint   `gorm:"index" json:"user_id"`
	Username string `gorm:"size:64" json:"username"`
	Action   string `gorm:"size:64;index;not null" json:"action"`     // login / create / update / delete / execute
	Resource string `gorm:"size:64;index" json:"resource"`           // target / session / user / config
	ResourceID uint  `json:"resource_id,omitempty"`
	Detail   string `gorm:"type:text" json:"detail"`                  // 操作详情
	IP       string `gorm:"size:64" json:"ip"`
	Status   string `gorm:"size:16" json:"status"`                   // success / failed
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
