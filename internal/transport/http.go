package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPTransport HTTP 传输实现
type HTTPTransport struct {
	client *http.Client
}

// NewHTTPTransport 创建 HTTP 传输实例
func NewHTTPTransport() *HTTPTransport {
	return &HTTPTransport{
		client: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// Request 发送 HTTP 请求
func (t *HTTPTransport) Request(ctx context.Context, req *Request) (*Response, error) {
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	method := req.Method
	if method == "" {
		method = "POST"
	}

	var body io.Reader
	if req.Body != nil {
		body = strings.NewReader(string(req.Body))
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 收集响应头
	headers := make(map[string]string)
	for k := range resp.Header {
		headers[k] = resp.Header.Get(k)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       raw,
		Duration:   duration,
	}, nil
}

// PostForm 发送表单 POST 请求（便捷方法）
func (t *HTTPTransport) PostForm(ctx context.Context, targetURL string, form url.Values) (*Response, error) {
	req := &Request{
		Method:  "POST",
		URL:     targetURL,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(form.Encode()),
	}
	return t.Request(ctx, req)
}
