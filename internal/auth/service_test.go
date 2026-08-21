package auth

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	if hash == "secret123" {
		t.Error("哈希不应与明文相同")
	}
	if !VerifyPassword(hash, "secret123") {
		t.Error("正确密码应验证通过")
	}
	if VerifyPassword(hash, "wrong") {
		t.Error("错误密码不应验证通过")
	}
}

func TestJWTIssueAndParse(t *testing.T) {
	svc := NewService("test-secret-key", time.Hour)

	token, err := svc.issueToken(&User{ID: 1, Username: "admin"})
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("解析 token 失败: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("期望 UserID=1，得到 %d", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("期望 Username=admin，得到 %q", claims.Username)
	}
}

func TestParseInvalidToken(t *testing.T) {
	svc := NewService("test-secret-key", time.Hour)
	if _, err := svc.ParseToken("invalid.token.string"); err == nil {
		t.Error("非法 token 应报错")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("127.0.0.1:admin") {
			t.Fatalf("第 %d 次尝试应被允许", i+1)
		}
	}
	if rl.Allow("127.0.0.1:admin") {
		t.Error("超过限流次数后应被拒绝")
	}

	// 其他 key 不受影响。
	if !rl.Allow("127.0.0.1:other") {
		t.Error("其他 key 不应被限流")
	}

	rl.Reset("127.0.0.1:admin")
	if !rl.Allow("127.0.0.1:admin") {
		t.Error("Reset 后应重新允许")
	}
}

func TestHasPermission(t *testing.T) {
	u := &User{
		Roles: []Role{
			{
				Name: RoleAdmin,
				Permissions: []Permission{
					{Code: PermTargetRead},
					{Code: PermTargetDelete},
				},
			},
		},
	}

	if !u.HasPermission(PermTargetRead) {
		t.Error("应拥有 target:read 权限")
	}
	if u.HasPermission(PermUserManage) {
		t.Error("不应拥有 user:manage 权限")
	}
}
