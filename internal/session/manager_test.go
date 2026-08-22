package session

import (
	"testing"
	"time"

	"github.com/webshell-manager/webshell-manager/internal/target"
)

func TestCreateAndGet(t *testing.T) {
	m := NewManager(5, time.Minute)

	s, err := m.Create(1, &target.Target{ID: 1, URL: "http://example.com"})
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if s.ID == "" {
		t.Error("会话 ID 不应为空")
	}

	got, err := m.Get(s.ID)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got.TargetID != 1 {
		t.Errorf("期望 TargetID=1，得到 %d", got.TargetID)
	}
}

func TestGetNotFound(t *testing.T) {
	m := NewManager(5, time.Minute)
	if _, err := m.Get("nonexistent"); err == nil {
		t.Error("获取不存在的会话应报错")
	}
}

func TestClose(t *testing.T) {
	m := NewManager(5, time.Minute)
	s, _ := m.Create(1, &target.Target{ID: 1})
	m.Close(s.ID)
	if _, err := m.Get(s.ID); err == nil {
		t.Error("关闭后获取会话应报错")
	}
}

func TestMaxPerUser(t *testing.T) {
	m := NewManager(2, time.Minute)

	if _, err := m.Create(1, &target.Target{ID: 1}); err != nil {
		t.Fatalf("第 1 个会话创建失败: %v", err)
	}
	if _, err := m.Create(1, &target.Target{ID: 2}); err != nil {
		t.Fatalf("第 2 个会话创建失败: %v", err)
	}
	if _, err := m.Create(1, &target.Target{ID: 3}); err == nil {
		t.Error("超过单用户会话上限应报错")
	}
}

func TestExpired(t *testing.T) {
	m := NewManager(5, 10*time.Millisecond)
	s, _ := m.Create(1, &target.Target{ID: 1})

	time.Sleep(20 * time.Millisecond)
	if _, err := m.Get(s.ID); err == nil {
		t.Error("超时会话应被回收")
	}
}

func TestCleanupExpired(t *testing.T) {
	m := NewManager(5, 10*time.Millisecond)
	_, _ = m.Create(1, &target.Target{ID: 1})
	_, _ = m.Create(1, &target.Target{ID: 2})

	time.Sleep(20 * time.Millisecond)
	removed := m.CleanupExpired()
	if removed != 2 {
		t.Errorf("期望清理 2 个会话，得到 %d", removed)
	}
}
