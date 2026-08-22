package model

// Target 目标模型（WebShell 节点）
type Target struct {
	BaseModel
	Name     string `gorm:"size:128;not null" json:"name"`
	URL      string `gorm:"size:512;not null" json:"url"`
	Type     string `gorm:"size:16;not null" json:"type"` // php / jsp / asp / aspx
	Password string `gorm:"size:256" json:"-"`           // AES-256-GCM 加密存储
	Encoding string `gorm:"size:32" json:"encoding"`     // 编码: auto/utf-8/gbk
	Remark   string `gorm:"size:512" json:"remark"`
	Status   string `gorm:"size:16" json:"status"` // online / offline / unknown
	// 高级配置
	Method   string `gorm:"size:16" json:"method"`     // GET / POST
	Headers  string `gorm:"size:2048" json:"headers"`  // JSON 格式自定义头
	Cookies  string `gorm:"size:2048" json:"cookies"`  // Cookie 字符串
	Timeout  int    `json:"timeout"`                   // 超时秒数
}

func (Target) TableName() string {
	return "targets"
}
