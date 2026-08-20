package model

import (
	"sync"
	"time"
)

// Session 表示一个终端会话
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool     `json:"active"`
	mu        sync.Mutex
}

// SessionManager 管理所有终端会话
type SessionManager struct {
	sessions sync.Map // map[string]*Session
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// Add 添加一个会话
func (sm *SessionManager) Add(s *Session) {
	sm.sessions.Store(s.ID, s)
}

// Get 获取一个会话
func (sm *SessionManager) Get(id string) (*Session, bool) {
	v, ok := sm.sessions.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Session), true
}

// Remove 删除一个会话
func (sm *SessionManager) Remove(id string) {
	sm.sessions.Delete(id)
}

// List 列出所有会话
func (sm *SessionManager) List() []*Session {
	var list []*Session
	sm.sessions.Range(func(key, value any) bool {
		s := value.(*Session)
		list = append(list, s)
		return true
	})
	return list
}

// SetActive 设置会话活跃状态
func (s *Session) SetActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = active
}

// IsActive 检查会话是否活跃
func (s *Session) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Active
}
