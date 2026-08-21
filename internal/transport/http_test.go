package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPTransportRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	tr := NewHTTPTransport(5*time.Second, true)
	resp, err := tr.Request(context.Background(), &Request{
		Method: http.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("Request 失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望 200，得到 %d", resp.StatusCode)
	}
	if string(resp.Body) != "hello" {
		t.Errorf("期望 body=hello，得到 %q", string(resp.Body))
	}
	if resp.Headers["X-Test"] != "ok" {
		t.Errorf("期望 X-Test=ok，得到 %q", resp.Headers["X-Test"])
	}
	if resp.Duration <= 0 {
		t.Error("Duration 应大于 0")
	}
}

func TestValidateURLRejectsNonHTTP(t *testing.T) {
	tr := NewHTTPTransport(5*time.Second, true)

	if err := tr.validateURL("ftp://example.com/x"); err == nil {
		t.Error("ftp 协议应被拒绝")
	}
	if err := tr.validateURL("file:///etc/passwd"); err == nil {
		t.Error("file 协议应被拒绝")
	}
	if err := tr.validateURL("http:///no-host"); err == nil {
		t.Error("缺少主机名应被拒绝")
	}
}

func TestValidateURLBlocksPrivateWhenNotAllowed(t *testing.T) {
	tr := NewHTTPTransport(5*time.Second, false)

	// 回环地址应被拒绝。
	if err := tr.validateURL("http://127.0.0.1:8080/x"); err == nil {
		t.Error("allowLocal=false 时应拒绝回环地址")
	}
	// 内网地址应被拒绝。
	if err := tr.validateURL("http://192.168.1.1/x"); err == nil {
		t.Error("allowLocal=false 时应拒绝内网地址")
	}
}

func TestValidateURLAllowsLocalWhenAllowed(t *testing.T) {
	tr := NewHTTPTransport(5*time.Second, true)

	if err := tr.validateURL("http://127.0.0.1:8080/x"); err != nil {
		t.Errorf("allowLocal=true 时应允许回环地址: %v", err)
	}
}
