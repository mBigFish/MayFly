package transport

import (
	"context"
	"time"
)

// Request 传输层请求
type Request struct {
	Method  string            // HTTP 方法
	URL     string            // 目标 URL
	Headers map[string]string // 请求头
	Body    []byte            // 请求体
	Timeout time.Duration     // 超时时间
}

// Response 传输层响应
type Response struct {
	StatusCode int               // HTTP 状态码
	Headers    map[string]string // 响应头
	Body       []byte            // 响应体
	Duration   time.Duration     // 耗时
}

// Transport 传输层接口
type Transport interface {
	Request(ctx context.Context, req *Request) (*Response, error)
}
