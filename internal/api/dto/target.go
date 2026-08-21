package dto

// CreateTargetRequest 创建目标请求。
type CreateTargetRequest struct {
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
	Method   string `json:"method"`
	Headers  string `json:"headers"`
	Cookies  string `json:"cookies"`
	Timeout  int    `json:"timeout"`
	Proxy    string `json:"proxy"`
	Encoding string `json:"encoding"`
	Remark   string `json:"remark"`
	GroupID  uint   `json:"group_id"`
}

// UpdateTargetRequest 更新目标请求。
type UpdateTargetRequest struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
	Method   string `json:"method"`
	Headers  string `json:"headers"`
	Cookies  string `json:"cookies"`
	Timeout  int    `json:"timeout"`
	Proxy    string `json:"proxy"`
	Encoding string `json:"encoding"`
	Remark   string `json:"remark"`
	GroupID  uint   `json:"group_id"`
}
