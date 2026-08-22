package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestBuildCommand(t *testing.T) {
	a := &PHPAdapter{}

	cases := []struct {
		op      *operation.Operation
		wantSub string
	}{
		{&operation.Operation{Type: operation.OperationListDir, Params: map[string]any{"path": "/tmp"}}, "ls -la"},
		{&operation.Operation{Type: operation.OperationReadFile, Params: map[string]any{"path": "/etc/passwd"}}, "cat"},
		{&operation.Operation{Type: operation.OperationWriteFile, Params: map[string]any{"path": "/tmp/x", "content": "hi"}}, "> "},
		{&operation.Operation{Type: operation.OperationRename, Params: map[string]any{"from": "a", "to": "b"}}, "mv"},
		{&operation.Operation{Type: operation.OperationMkdir, Params: map[string]any{"path": "d"}}, "mkdir"},
		{&operation.Operation{Type: operation.OperationDelete, Params: map[string]any{"path": "d"}}, "rm -rf"},
		{&operation.Operation{Type: operation.OperationSystemInfo}, "uname"},
	}

	for _, c := range cases {
		cmd, err := a.buildCommand(c.op)
		if err != nil {
			t.Errorf("buildCommand(%s) 失败: %v", c.op.Type, err)
			continue
		}
		if !strings.Contains(cmd, c.wantSub) {
			t.Errorf("buildCommand(%s) = %q，期望包含 %q", c.op.Type, cmd, c.wantSub)
		}
	}
}

func TestBuildCommandMissingParam(t *testing.T) {
	a := &PHPAdapter{}
	if _, err := a.buildCommand(&operation.Operation{Type: operation.OperationReadFile}); err == nil {
		t.Error("缺少 path 应报错")
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
