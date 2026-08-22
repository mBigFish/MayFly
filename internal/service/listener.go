package service

import (
	"fmt"
	"net"
	"sync"

	"gorm.io/gorm"
	"mayfly/internal/model"
)

// ListenerManager 监听器管理器
type ListenerManager struct {
	mu        sync.RWMutex
	listeners map[uint]*listenerEntry
}

type listenerEntry struct {
	model     *model.Listener
	tcpListener net.Listener
	clients   map[string]net.Conn
}

var (
	listenerMgr     *ListenerManager
	listenerMgrOnce sync.Once
)

// GetListenerManager 获取全局监听器管理器
func GetListenerManager() *ListenerManager {
	listenerMgrOnce.Do(func() {
		listenerMgr = &ListenerManager{
			listeners: make(map[uint]*listenerEntry),
		}
	})
	return listenerMgr
}

// ListListeners 查询监听器列表
func ListListeners(db *gorm.DB) ([]model.Listener, error) {
	var listeners []model.Listener
	err := db.Order("id DESC").Find(&listeners).Error
	// 同步运行状态
	mgr := GetListenerManager()
	for i := range listeners {
		if _, ok := mgr.listeners[listeners[i].ID]; ok {
			listeners[i].Status = "running"
		} else {
			listeners[i].Status = "stopped"
		}
	}
	return listeners, err
}

// CreateListener 创建监听器记录
func CreateListener(db *gorm.DB, l *model.Listener) error {
	return db.Create(l).Error
}

// StartListener 启动监听器
func (m *ListenerManager) StartListener(db *gorm.DB, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.listeners[id]; ok {
		return fmt.Errorf("监听器已在运行")
	}

	var l model.Listener
	if err := db.First(&l, id).Error; err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", l.Host, l.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听失败: %w", err)
	}

	entry := &listenerEntry{
		model:     &l,
		tcpListener: ln,
		clients:   make(map[string]net.Conn),
	}
	m.listeners[id] = entry

	// 接受连接的 goroutine
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			addr := conn.RemoteAddr().String()
			entry.clients[addr] = conn
			// 更新连接数
			db.Exec("UPDATE listeners SET connections = connections + 1 WHERE id = ?", id)
		}
	}()

	return nil
}

// StopListener 停止监听器
func (m *ListenerManager) StopListener(db *gorm.DB, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.listeners[id]
	if !ok {
		return fmt.Errorf("监听器未运行")
	}

	if entry.tcpListener != nil {
		entry.tcpListener.Close()
	}
	for _, conn := range entry.clients {
		conn.Close()
	}
	delete(m.listeners, id)

	return nil
}

// DeleteListener 删除监听器
func DeleteListener(db *gorm.DB, id uint) error {
	mgr := GetListenerManager()
	_ = mgr.StopListener(db, id)
	return db.Delete(&model.Listener{}, id).Error
}
