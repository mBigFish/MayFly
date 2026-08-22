// Package session 定义会话实体与生命周期管理。
package session

import (
	"time"

	"github.com/webshell-manager/webshell-manager/internal/target"
)

// Session 是一次终端/操作会话，绑定到授权目标。
type Session struct {
	ID        string    `json:"id"`
	TargetID  uint      `json:"target_id"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`

	// target 保存解密后的目标凭证，供命令执行使用（不序列化）。
	target *target.Target
}

// Target 返回会话绑定的目标（含解密后的敏感字段）。
func (s *Session) Target() *target.Target {
	return s.target
}

// Touch 更新最近活跃时间。
func (s *Session) Touch() {
	s.LastSeen = time.Now()
}

// Expired 判断会话是否已超时。
func (s *Session) Expired(timeout time.Duration) bool {
	return time.Since(s.LastSeen) > timeout
}
