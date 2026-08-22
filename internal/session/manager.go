// Package session 会话管理器：负责会话的创建、获取、关闭、列表与超时回收。
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/webshell-manager/webshell-manager/internal/target"
)

// Manager 会话管理器。
type Manager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	maxPerUser  int           // 单用户最大会话数
	idleTimeout time.Duration // 会话空闲超时
}

// NewManager 创建会话管理器。
func NewManager(maxPerUser int, idleTimeout time.Duration) *Manager {
	return &Manager{
		sessions:    make(map[string]*Session),
		maxPerUser:  maxPerUser,
		idleTimeout: idleTimeout,
	}
}

// Create 为用户和目标创建一个新会话。
func (m *Manager) Create(userID uint, t *target.Target) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 单用户会话数量限制。
	count := 0
	for _, s := range m.sessions {
		if s.UserID == userID {
			count++
		}
	}
	if count >= m.maxPerUser {
		return nil, fmt.Errorf("会话数量已达上限（%d）", m.maxPerUser)
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	s := &Session{
		ID:        id,
		TargetID:  t.ID,
		UserID:    userID,
		CreatedAt: now,
		LastSeen:  now,
		target:    t,
	}
	m.sessions[id] = s
	return s, nil
}

// Get 按 ID 获取会话，并刷新活跃时间。
func (m *Manager) Get(id string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("会话不存在")
	}
	if s.Expired(m.idleTimeout) {
		delete(m.sessions, id)
		return nil, fmt.Errorf("会话已超时")
	}
	s.Touch()
	return s, nil
}

// Touch 刷新会话活跃时间（不要求存在）。
func (m *Manager) Touch(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.Touch()
	}
}

// Close 关闭并删除会话。
func (m *Manager) Close(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

// List 返回指定用户的会话列表（副本）。
func (m *Manager) List(userID uint) []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Session, 0)
	for _, s := range m.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result
}

// CleanupExpired 回收所有已超时的会话。
func (m *Manager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for id, s := range m.sessions {
		if s.Expired(m.idleTimeout) {
			delete(m.sessions, id)
			removed++
		}
	}
	return removed
}

// newID 生成随机会话 ID。
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
