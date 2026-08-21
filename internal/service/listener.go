package service

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Listener 监听端口实例
type Listener struct {
	ID        string    `json:"id"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"` // tcp / udp
	Status    string    `json:"status"`   // listening / stopped / error
	CreatedAt time.Time `json:"created_at"`
	ln        net.Listener
	stop      chan struct{}
	output    string
	mu        sync.Mutex
}

// ListenerManager 管理所有监听端口
type ListenerManager struct {
	mu        sync.RWMutex
	listeners map[string]*Listener
}

// NewListenerManager 创建监听管理器
func NewListenerManager() *ListenerManager {
	return &ListenerManager{listeners: make(map[string]*Listener)}
}

// StartListener 启动一个监听端口
func (m *ListenerManager) StartListener(id string, port int, protocol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if old, ok := m.listeners[id]; ok && old.Status == "listening" {
		return fmt.Errorf("监听 %s 已在运行", id)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("监听端口 %d 失败: %v", port, err)
	}

	l := &Listener{
		ID:        id,
		Port:      port,
		Protocol:  protocol,
		Status:    "listening",
		CreatedAt: time.Now(),
		ln:        ln,
		stop:      make(chan struct{}),
	}

	m.listeners[id] = l

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-l.stop:
					return
				default:
					l.mu.Lock()
					l.output += fmt.Sprintf("[!] Accept error: %v\n", err)
					l.mu.Unlock()
					return
				}
			}
			go l.handleConn(conn)
		}
	}()

	return nil
}

func (l *Listener) handleConn(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	l.mu.Lock()
	l.output += fmt.Sprintf("[+] %s - New connection from %s\n", time.Now().Format("15:04:05"), addr)
	l.mu.Unlock()

	buf := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		n, err := conn.Read(buf)
		if n > 0 {
			l.mu.Lock()
			l.output += string(buf[:n])
			l.mu.Unlock()
		}
		if err != nil {
			l.mu.Lock()
			l.output += fmt.Sprintf("\n[-] %s - Connection %s closed\n", time.Now().Format("15:04:05"), addr)
			l.mu.Unlock()
			return
		}
	}
}

// StopListener 停止监听
func (m *ListenerManager) StopListener(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, ok := m.listeners[id]
	if !ok {
		return fmt.Errorf("监听 %s 不存在", id)
	}
	if l.Status == "stopped" {
		return nil
	}
	close(l.stop)
	l.ln.Close()
	l.Status = "stopped"
	return nil
}

// GetListener 获取监听信息
func (m *ListenerManager) GetListener(id string) (*Listener, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.listeners[id]
	return l, ok
}

// GetOutput 获取监听输出
func (m *ListenerManager) GetOutput(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.listeners[id]
	if !ok {
		return "", fmt.Errorf("监听 %s 不存在", id)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.output, nil
}

// ListListeners 列出所有监听
func (m *ListenerManager) ListListeners() []*Listener {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Listener, 0, len(m.listeners))
	for _, l := range m.listeners {
		list = append(list, l)
	}
	return list
}

// DeleteListener 删除监听
func (m *ListenerManager) DeleteListener(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.listeners[id]
	if !ok {
		return fmt.Errorf("监听 %s 不存在", id)
	}
	if l.Status == "listening" {
		close(l.stop)
		l.ln.Close()
	}
	delete(m.listeners, id)
	return nil
}
