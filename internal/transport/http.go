package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPTransport 是基于 net/http 的传输实现。
// 内置 SSRF 防护：仅允许 http/https，禁止内网地址（可配置），禁用自动重定向。
type HTTPTransport struct {
	client     *http.Client
	allowLocal bool // 是否允许访问内网/回环地址（仅用于本地授权测试服务）
}

// NewHTTPTransport 创建 HTTP 传输。
// timeout 为单次请求超时；allowLocal 为 true 时允许访问内网（用于本地测试服务）。
func NewHTTPTransport(timeout time.Duration, allowLocal bool) *HTTPTransport {
	return &HTTPTransport{
		client: &http.Client{
			Timeout: timeout,
			// 禁用自动重定向，由上层控制（SSRF 防护）。
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		allowLocal: allowLocal,
	}
}

// Request 执行一次 HTTP 请求。
func (t *HTTPTransport) Request(ctx context.Context, req *Request) (*Response, error) {
	// 校验 URL（SSRF 防护）。
	if err := t.validateURL(req.URL); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	if httpReq.Header.Get("Content-Type") == "" && len(req.Body) > 0 {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	start := time.Now()
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	headers := make(map[string]string, len(resp.Header))
	for k := range resp.Header {
		headers[k] = resp.Header.Get(k)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
		Duration:   time.Since(start),
	}, nil
}

// validateURL 校验 URL，防止 SSRF。
func (t *HTTPTransport) validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("URL 格式非法: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("仅允许 http/https 协议")
	}
	if u.Host == "" {
		return fmt.Errorf("URL 缺少主机名")
	}

	// 内网访问策略。
	if !t.allowLocal {
		if err := t.checkPrivateHost(u.Hostname()); err != nil {
			return err
		}
	}

	return nil
}

// checkPrivateHost 检查主机是否为内网/回环/保留地址。
func (t *HTTPTransport) checkPrivateHost(host string) error {
	ip := net.ParseIP(host)
	if ip == nil {
		// 域名：解析后逐一检查。
		addrs, err := net.LookupIP(host)
		if err != nil {
			return fmt.Errorf("解析主机失败: %w", err)
		}
		for _, addr := range addrs {
			if isPrivateIP(addr) {
				return fmt.Errorf("禁止访问内网地址: %s", host)
			}
		}
		return nil
	}
	if isPrivateIP(ip) {
		return fmt.Errorf("禁止访问内网地址: %s", host)
	}
	return nil
}

// isPrivateIP 判断 IP 是否为私网/回环/链路本地/保留地址。
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	// 未指定地址（0.0.0.0）。
	if ip.IsUnspecified() {
		return true
	}
	return false
}
