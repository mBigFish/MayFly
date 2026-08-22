package protocol

import (
	"context"
	"fmt"

	"mayfly/internal/model"
	"mayfly/internal/transport"
)

// OperationType 操作类型
type OperationType string

const (
	OpCommand     OperationType = "command"
	OpReadFile    OperationType = "read_file"
	OpListDir     OperationType = "list_dir"
	OpWriteFile   OperationType = "write_file"
	OpDeleteFile  OperationType = "delete_file"
	OpRenameFile  OperationType = "rename_file"
	OpMkdir       OperationType = "mkdir"
	OpSysInfo     OperationType = "system_info"
	OpDBQuery     OperationType = "db_query"
)

// Operation 操作描述
type Operation struct {
	Type   OperationType       `json:"type"`
	Params map[string]string   `json:"params"`
}

// Result 操作结果
type Result struct {
	Status  string `json:"status"`   // ok / error
	Data    []byte `json:"-"`        // 原始结果数据
	Message string `json:"message"`  // 错误信息
}

// FileEntry 文件条目
type FileEntry struct {
	Type  string `json:"type"`  // d / f
	Size  int64  `json:"size"`
	Mtime int64  `json:"mtime"`
	Name  string `json:"name"`
}

// FileList 目录列表结果
type FileList struct {
	Path    string       `json:"path"`
	Parent  string       `json:"parent"`
	Entries []*FileEntry `json:"entries"`
}

// Protocol 协议适配器接口
type Protocol interface {
	Name() string
	Check(ctx context.Context, target *model.Target) error
	Execute(ctx context.Context, target *model.Target, op *Operation) (*Result, error)
}

// transportHolder 持有传输层实例
type transportHolder struct {
	transport transport.Transport
}

// NewTransportHolder 创建传输层持有者
func NewTransportHolder(t transport.Transport) *transportHolder {
	return &transportHolder{transport: t}
}

// GetTransport 获取传输层
func (h *transportHolder) GetTransport() transport.Transport {
	return h.transport
}

// DecodeError 从结果中提取错误
func DecodeError(r *Result) error {
	if r == nil {
		return fmt.Errorf("nil result")
	}
	if r.Status != "ok" {
		return fmt.Errorf("%s", r.Message)
	}
	return nil
}
