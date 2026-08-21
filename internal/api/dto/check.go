package dto

// CheckResponse 目标探活响应。
type CheckResponse struct {
	TargetID uint   `json:"target_id"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}
