// Package transport 定义统一的网络传输抽象。
// 协议层必须通过 Transport 与目标通信，不得直接使用 http.Client。
package transport

import (
	"context"
	"time"
)

// Transport 是统一传输接口，屏蔽具体通信方式。
// 第一阶段实现 HTTP Transport，后续可扩展 Proxy / Custom Transport。
type Transport interface {
	Request(ctx context.Context, req *Request) (*Response, error)
}

// Request 是一次传输请求。
type Request struct {
	Method  string            // HTTP 方法
	URL     string            // 目标地址
	Headers map[string]string // 请求头
	Body    []byte            // 请求体
}

// Response 是一次传输响应。
type Response struct {
	StatusCode int               // 状态码
	Headers    map[string]string // 响应头
	Body       []byte            // 响应体
	Duration   time.Duration     // 请求耗时
}
