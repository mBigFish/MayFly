package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mayfly/internal/model"
	"mayfly/internal/plugin"
	"mayfly/internal/protocol"
)

// SystemInfoPlugin 系统信息插件
type SystemInfoPlugin struct{}

func (p *SystemInfoPlugin) Name() string        { return "system_info" }
func (p *SystemInfoPlugin) Version() string     { return "1.0.0" }
func (p *SystemInfoPlugin) Description() string { return "获取目标系统信息（OS、用户、IP、主机名等）" }

func (p *SystemInfoPlugin) Execute(ctx context.Context, target *model.Target, params map[string]string) (*plugin.Result, error) {
	result, err := plugin.ExecuteViaProtocol(ctx, target, protocol.OpSysInfo, params)
	if err != nil {
		return &plugin.Result{Status: "error", Message: err.Error()}, err
	}

	// 尝试解析 JSON
	var info interface{}
	if err := json.Unmarshal(result.Data, &info); err == nil {
		return &plugin.Result{Status: "ok", Data: info}, nil
	}

	return &plugin.Result{
		Status:  result.Status,
		Data:    string(result.Data),
		Message: result.Message,
	}, nil
}

// ProcessViewerPlugin 进程查看插件
type ProcessViewerPlugin struct{}

func (p *ProcessViewerPlugin) Name() string        { return "process_viewer" }
func (p *ProcessViewerPlugin) Version() string     { return "1.0.0" }
func (p *ProcessViewerPlugin) Description() string { return "查看目标进程列表（通过命令执行）" }

func (p *ProcessViewerPlugin) Execute(ctx context.Context, target *model.Target, params map[string]string) (*plugin.Result, error) {
	cmd := params["cmd"]
	if cmd == "" {
		// 根据目标类型选择命令
		if strings.Contains(target.URL, ".aspx") {
			cmd = "tasklist"
		} else if strings.Contains(target.URL, ".asp") {
			cmd = "tasklist"
		} else {
			cmd = "ps aux"
		}
	}

	result, err := plugin.ExecuteViaProtocol(ctx, target, protocol.OpCommand, map[string]string{"cmd": cmd})
	if err != nil {
		return &plugin.Result{Status: "error", Message: err.Error()}, err
	}

	return &plugin.Result{
		Status:  result.Status,
		Data:    string(result.Data),
		Message: result.Message,
	}, nil
}

// NetworkInfoPlugin 网络信息插件
type NetworkInfoPlugin struct{}

func (p *NetworkInfoPlugin) Name() string        { return "network_info" }
func (p *NetworkInfoPlugin) Version() string     { return "1.0.0" }
func (p *NetworkInfoPlugin) Description() string { return "查看目标网络连接信息" }

func (p *NetworkInfoPlugin) Execute(ctx context.Context, target *model.Target, params map[string]string) (*plugin.Result, error) {
	cmd := params["cmd"]
	if cmd == "" {
		if strings.Contains(target.URL, ".aspx") || strings.Contains(target.URL, ".asp") {
			cmd = "netstat -an"
		} else {
			cmd = "netstat -tulpn"
		}
	}

	result, err := plugin.ExecuteViaProtocol(ctx, target, protocol.OpCommand, map[string]string{"cmd": cmd})
	if err != nil {
		return &plugin.Result{Status: "error", Message: err.Error()}, err
	}

	return &plugin.Result{
		Status:  result.Status,
		Data:    string(result.Data),
		Message: result.Message,
	}, nil
}

// RegisterAll 注册所有内置插件
func RegisterAll() {
	m := plugin.GetManager()
	m.Register(&SystemInfoPlugin{})
	m.Register(&ProcessViewerPlugin{})
	m.Register(&NetworkInfoPlugin{})
}

// 确保 fmt 被使用
var _ = fmt.Sprintf
