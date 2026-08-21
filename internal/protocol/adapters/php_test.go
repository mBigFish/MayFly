package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/webshell-manager/webshell-manager/internal/operation"
	"github.com/webshell-manager/webshell-manager/internal/target"
	"github.com/webshell-manager/webshell-manager/internal/transport"
)

func newPHPAdapterWithServer(t *testing.T, handler http.HandlerFunc) (*PHPAdapter, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tr := transport.NewHTTPTransport(5*time.Second, true)
	return NewPHPAdapter(tr), srv
}

func TestPHPCheckSuccess(t *testing.T) {
	adapter, srv := newPHPAdapterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("cmd") == "echo WSM_OK" {
			_, _ = w.Write([]byte("WSM_OK"))
		}
	})

	tg := &target.Target{URL: srv.URL, Method: "POST", Protocol: "php"}
	if err := adapter.Check(context.Background(), tg); err != nil {
		t.Errorf("Check 应成功: %v", err)
	}
}

func TestPHPCheckFail(t *testing.T) {
	adapter, srv := newPHPAdapterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not a webshell"))
	})

	tg := &target.Target{URL: srv.URL, Method: "POST", Protocol: "php"}
	if err := adapter.Check(context.Background(), tg); err == nil {
		t.Error("非 webshell 响应应使 Check 失败")
	}
}

func TestPHPExecuteCommand(t *testing.T) {
	adapter, srv := newPHPAdapterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = w.Write([]byte("executed: " + r.FormValue("cmd")))
	})

	tg := &target.Target{URL: srv.URL, Method: "POST", Protocol: "php"}
	res, err := adapter.Execute(context.Background(), tg, &operation.Operation{
		Type:   operation.OperationCommand,
		Params: map[string]any{"cmd": "id"},
	})
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}
	if !res.Success {
		t.Errorf("期望 Success=true，得到 %v", res.Success)
	}
	if res.Output != "executed: id" {
		t.Errorf("期望 output=executed: id，得到 %q", res.Output)
	}
}

func TestParseHeaderString(t *testing.T) {
	h := parseHeaderString("Authorization: Bearer x\nX-Custom: y\n")
	if h["Authorization"] != "Bearer x" {
		t.Errorf("Authorization 解析错误: %q", h["Authorization"])
	}
	if h["X-Custom"] != "y" {
		t.Errorf("X-Custom 解析错误: %q", h["X-Custom"])
	}
}
