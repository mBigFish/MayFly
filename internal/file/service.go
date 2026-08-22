// Package file 实现文件管理服务，通过协议层执行文件操作。
// 所有路径操作都必须进行规范化与合法性校验（spec 第 23 节）。
package file

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/webshell-manager/webshell-manager/internal/operation"
	"github.com/webshell-manager/webshell-manager/internal/protocol"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// Service 文件管理服务。
type Service struct {
	registry *protocol.Registry
}

// NewService 创建文件管理服务。
func NewService(registry *protocol.Registry) *Service {
	return &Service{registry: registry}
}

// normalizePath 规范化并校验路径，防止路径穿越。
func normalizePath(p string) (string, error) {
	if p == "" {
		return ".", nil
	}
	// 拒绝绝对路径之外的非法字符。
	if strings.ContainsRune(p, '\x00') {
		return "", fmt.Errorf("路径包含非法字符")
	}
	clean := path.Clean(p)
	// 阻止穿越到上级目录（如 ../ 已通过 Clean 归一，但保留对根外逃逸的防护）。
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("路径不允许越界")
	}
	return clean, nil
}

// getProtocol 根据目标协议获取协议适配器。
func (s *Service) getProtocol(t *target.Target) (protocol.Protocol, error) {
	return s.registry.Get(t.Protocol)
}

// List 列出目录内容。
func (s *Service) List(ctx context.Context, t *target.Target, dir string) (*operation.Result, error) {
	p, err := normalizePath(dir)
	if err != nil {
		return nil, err
	}
	proto, err := s.getProtocol(t)
	if err != nil {
		return nil, err
	}
	return proto.Execute(ctx, t, &operation.Operation{
		Type:   operation.OperationListDir,
		Params: map[string]any{"path": p},
	})
}

// Read 读取文件内容。
func (s *Service) Read(ctx context.Context, t *target.Target, path string) (*operation.Result, error) {
	p, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	proto, err := s.getProtocol(t)
	if err != nil {
		return nil, err
	}
	return proto.Execute(ctx, t, &operation.Operation{
		Type:   operation.OperationReadFile,
		Params: map[string]any{"path": p},
	})
}

// Write 写入文件内容。
func (s *Service) Write(ctx context.Context, t *target.Target, path, content string) (*operation.Result, error) {
	p, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	proto, err := s.getProtocol(t)
	if err != nil {
		return nil, err
	}
	return proto.Execute(ctx, t, &operation.Operation{
		Type:   operation.OperationWriteFile,
		Params: map[string]any{"path": p, "content": content},
	})
}

// Rename 重命名/移动文件。
func (s *Service) Rename(ctx context.Context, t *target.Target, from, to string) (*operation.Result, error) {
	f, err := normalizePath(from)
	if err != nil {
		return nil, err
	}
	tt, err := normalizePath(to)
	if err != nil {
		return nil, err
	}
	proto, err := s.getProtocol(t)
	if err != nil {
		return nil, err
	}
	return proto.Execute(ctx, t, &operation.Operation{
		Type:   operation.OperationRename,
		Params: map[string]any{"from": f, "to": tt},
	})
}

// Mkdir 创建目录。
func (s *Service) Mkdir(ctx context.Context, t *target.Target, dir string) (*operation.Result, error) {
	p, err := normalizePath(dir)
	if err != nil {
		return nil, err
	}
	proto, err := s.getProtocol(t)
	if err != nil {
		return nil, err
	}
	return proto.Execute(ctx, t, &operation.Operation{
		Type:   operation.OperationMkdir,
		Params: map[string]any{"path": p},
	})
}

// Delete 删除文件或目录。
func (s *Service) Delete(ctx context.Context, t *target.Target, path string) (*operation.Result, error) {
	p, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	proto, err := s.getProtocol(t)
	if err != nil {
		return nil, err
	}
	return proto.Execute(ctx, t, &operation.Operation{
		Type:   operation.OperationDelete,
		Params: map[string]any{"path": p},
	})
}
