package auth

import (
	"sync"
	"time"
)

// RateLimiter 简单的内存滑动窗口限流器，用于登录接口防暴力破解。
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter 创建限流器。limit 为 window 时间窗口内的最大尝试次数。
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow 判断指定 key 是否允许继续尝试。
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// 清理过期记录。
	var valid []time.Time
	for _, t := range r.attempts[key] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= r.limit {
		// 该 key 已被限流，保留计数以便持续拒绝，直到窗口滑过。
		r.attempts[key] = valid
		return false
	}

	valid = append(valid, now)
	r.attempts[key] = valid

	// 清理无记录的 key，避免内存无限增长。
	if len(valid) == 0 {
		delete(r.attempts, key)
	}
	return true
}

// Reset 清空指定 key 的计数（登录成功后调用）。
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attempts, key)
}
