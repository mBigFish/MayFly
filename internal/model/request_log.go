package model

// RequestLog 请求日志（Request Inspector）
type RequestLog struct {
	BaseModel
	TargetID   uint   `json:"target_id" gorm:"index"`
	TargetName string `json:"target_name"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Operation  string `json:"operation"`  // command, read_file, list_dir, etc.
	Params     string `json:"params"`     // JSON encoded params
	Request    string `json:"request"`    // 请求体
	Response   string `json:"response"`   // 响应体
	Status     string `json:"status"`     // ok / error
	Duration   int64  `json:"duration"`   // 毫秒
	Error      string `json:"error"`
}
