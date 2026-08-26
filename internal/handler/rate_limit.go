package handler

import (
	"sync"
	"time"
)

// 登录防爆破配置
const (
	// maxLoginFailures 连续失败次数达到该值后锁定
	maxLoginFailures = 5
	// loginLockDuration 锁定持续时间
	loginLockDuration = 15 * time.Minute
	// loginFailDelay 登录失败时的响应延迟，减缓爆破速度
	loginFailDelay = 900 * time.Millisecond
)

// loginAttempt 单个 IP 的登录尝试记录
type loginAttempt struct {
	failCount   int
	lockedUntil time.Time
}

// loginLimiter 基于 IP 的登录限流器（内存实现，进程重启后重置）
type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

var limiter = &loginLimiter{
	attempts: make(map[string]*loginAttempt),
}

// isLocked 返回该 IP 当前是否处于锁定状态，同时清理已过期的记录
func (l *loginLimiter) isLocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[ip]
	if !ok {
		return false
	}
	// 尚未触发锁定（还在累计失败次数阶段），不删除记录
	if a.lockedUntil.IsZero() {
		return false
	}
	// 锁定期已过则清除记录
	if time.Now().After(a.lockedUntil) {
		delete(l.attempts, ip)
		return false
	}
	return true
}

// recordFailure 记录一次失败，失败次数达到阈值后进入锁定
func (l *loginLimiter) recordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[ip]
	if !ok {
		a = &loginAttempt{}
		l.attempts[ip] = a
	}
	a.failCount++
	if a.failCount >= maxLoginFailures {
		a.lockedUntil = time.Now().Add(loginLockDuration)
		a.failCount = 0
	}
}

// reset 登录成功后清除该 IP 的失败记录
func (l *loginLimiter) reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

// getRemaining 返回该 IP 剩余的失败次数（到达阈值后锁定）
func (l *loginLimiter) getRemaining(ip string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	a, ok := l.attempts[ip]
	if !ok {
		return maxLoginFailures
	}
	remain := maxLoginFailures - a.failCount
	if remain < 0 {
		remain = 0
	}
	return remain
}
