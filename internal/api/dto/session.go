package dto

// CreateSessionRequest 创建会话请求。
type CreateSessionRequest struct {
	TargetID uint `json:"target_id" binding:"required"`
}

// SessionResponse 会话响应。
type SessionResponse struct {
	ID        string `json:"id"`
	TargetID  uint   `json:"target_id"`
	UserID    uint   `json:"user_id"`
	CreatedAt string `json:"created_at"`
	LastSeen  string `json:"last_seen"`
}
