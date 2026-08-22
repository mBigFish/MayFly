package plugin

import (
	"context"
	"fmt"
	"sync"

	"mayfly/internal/model"
	"mayfly/internal/protocol"
)

// Manager 插件管理器
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

var (
	manager     *Manager
	managerOnce sync.Once
)

// GetManager 获取全局插件管理器
func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			plugins: make(map[string]Plugin),
		}
	})
	return manager
}

// Register 注册插件
func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.Name()] = p
}

// Get 获取插件
func (m *Manager) Get(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", name)
	}
	return p, nil
}

// List 列出所有插件
func (m *Manager) List() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		list = append(list, p)
	}
	return list
}

// Execute 执行插件
func (m *Manager) Execute(ctx context.Context, name string, target *model.Target, params map[string]string) (*Result, error) {
	p, err := m.Get(name)
	if err != nil {
		return nil, err
	}
	return p.Execute(ctx, target, params)
}

// ExecuteViaProtocol 通过协议适配器执行操作（内置插件辅助函数）
func ExecuteViaProtocol(ctx context.Context, target *model.Target, opType protocol.OperationType, params map[string]string) (*protocol.Result, error) {
	return protocol.ExecuteForTarget(ctx, target, &protocol.Operation{
		Type:   opType,
		Params: params,
	})
}
