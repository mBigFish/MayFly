// Package dto 定义 API 层的数据传输对象与统一响应格式。
package dto

// Response 统一 API 响应结构（对应开发规则第 19 条）。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(data interface{}) Response {
	return Response{Code: 0, Message: "ok", Data: data}
}

// Error 返回错误响应。
func Error(code int, message string) Response {
	return Response{Code: code, Message: message}
}

// PageData 分页数据。
type PageData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}
