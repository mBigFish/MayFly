package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"mayfly/internal/model"
)

// SessionService 会话服务
type SessionService struct {
	manager    *model.SessionManager
	ptySessions sync.Map // map[string]Terminal
}

// NewSessionService 创建会话服务
func NewSessionService() *SessionService {
	return &SessionService{
		manager: model.NewSessionManager(),
	}
}

// CreateSession 创建一个新的终端会话
func (ss *SessionService) CreateSession(name string) (*model.Session, error) {
	id := generateID()
	if name == "" {
		name = fmt.Sprintf("Terminal %s", id[:8])
	}

	session := &model.Session{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		Active:    true,
	}

	ss.manager.Add(session)
	return session, nil
}

// GetSession 获取会话
func (ss *SessionService) GetSession(id string) (*model.Session, bool) {
	return ss.manager.Get(id)
}

// ListSessions 列出所有会话
func (ss *SessionService) ListSessions() []*model.Session {
	return ss.manager.List()
}

// RemoveSession 删除会话
func (ss *SessionService) RemoveSession(id string) {
	ss.manager.Remove(id)
	if pty, ok := ss.ptySessions.Load(id); ok {
		pty.(Terminal).Close()
		ss.ptySessions.Delete(id)
	}
}

// StorePTY 存储会话对应的终端
func (ss *SessionService) StorePTY(sessionID string, t Terminal) {
	ss.ptySessions.Store(sessionID, t)
}

// GetPTY 获取会话对应的终端
func (ss *SessionService) GetPTY(sessionID string) (Terminal, bool) {
	v, ok := ss.ptySessions.Load(sessionID)
	if !ok {
		return nil, false
	}
	return v.(Terminal), true
}

// generateID 生成随机会话 ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
