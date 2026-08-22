package protocol

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"mayfly/internal/model"
)

// Registry 协议注册中心
type Registry struct {
	mu        sync.RWMutex
	protocols map[string]Protocol
}

var (
	registry     *Registry
	registryOnce sync.Once
)

// GetRegistry 获取全局注册中心
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		registry = &Registry{
			protocols: make(map[string]Protocol),
		}
	})
	return registry
}

// Register 注册协议适配器
func (r *Registry) Register(p Protocol) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.protocols[strings.ToLower(p.Name())] = p
}

// Get 获取协议适配器
func (r *Registry) Get(name string) (Protocol, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.protocols[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("不支持的协议类型: %s", name)
	}
	return p, nil
}

// List 列出所有已注册协议
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.protocols))
	for name := range r.protocols {
		names = append(names, name)
	}
	return names
}

// ExecuteForTarget 为指定目标执行操作
func ExecuteForTarget(ctx context.Context, target *model.Target, op *Operation) (*Result, error) {
	r := GetRegistry()
	p, err := r.Get(target.Type)
	if err != nil {
		return nil, err
	}
	return p.Execute(ctx, target, op)
}

// CheckTarget 检查目标连接
func CheckTarget(ctx context.Context, target *model.Target) error {
	r := GetRegistry()
	p, err := r.Get(target.Type)
	if err != nil {
		return err
	}
	return p.Check(ctx, target)
}
