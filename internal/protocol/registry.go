package protocol

import (
	"fmt"
	"sync"
)

// Registry 管理协议实例的注册与查找。
type Registry struct {
	mu        sync.RWMutex
	protocols map[string]Protocol
}

// NewRegistry 创建协议注册表。
func NewRegistry() *Registry {
	return &Registry{protocols: make(map[string]Protocol)}
}

// Register 注册协议。
func (r *Registry) Register(p Protocol) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.protocols[p.Name()] = p
}

// Get 按名称获取协议。
func (r *Registry) Get(name string) (Protocol, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.protocols[name]
	if !ok {
		return nil, fmt.Errorf("不支持的协议: %q", name)
	}
	return p, nil
}

// Names 返回所有已注册协议名称。
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.protocols))
	for name := range r.protocols {
		names = append(names, name)
	}
	return names
}
