package target

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/webshell-manager/webshell-manager/internal/config"
	"github.com/webshell-manager/webshell-manager/internal/crypto"
	"github.com/webshell-manager/webshell-manager/internal/database"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := database.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	// 建表。
	if err := db.AutoMigrate(&Target{}, &TargetGroup{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	encrypt, err := crypto.NewAESGCM([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("创建加密器失败: %v", err)
	}

	repo := NewRepository(db)
	return NewService(repo, encrypt)
}

func TestValidateURL(t *testing.T) {
	svc := newTestService(t)

	// 合法 http。
	ok := &Target{Name: "t", URL: "http://example.com/shell.php"}
	if err := svc.validate(ok); err != nil {
		t.Errorf("合法 http URL 应通过: %v", err)
	}

	// 非法协议。
	bad := &Target{Name: "t", URL: "ftp://example.com/x"}
	if err := svc.validate(bad); err == nil {
		t.Error("非 http/https 协议应报错")
	}

	// 缺少主机。
	nohost := &Target{Name: "t", URL: "http:///path"}
	if err := svc.validate(nohost); err == nil {
		t.Error("缺少主机名应报错")
	}

	// 空名称。
	noname := &Target{URL: "http://example.com"}
	if err := svc.validate(noname); err == nil {
		t.Error("空名称应报错")
	}
}

func TestCreateEncryptsSensitiveFields(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	tg := &Target{
		Name:    "test-target",
		URL:     "http://example.com/shell.php",
		Cookies: "PHPSESSID=secret",
		Headers: "Authorization: Bearer secret",
	}
	if err := svc.Create(ctx, tg); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	// 数据库中存储的应是密文。
	raw, err := svc.repo.GetByID(ctx, tg.ID)
	if err != nil {
		t.Fatalf("读取原始记录失败: %v", err)
	}
	if raw.Cookies == "PHPSESSID=secret" {
		t.Error("Cookies 应以密文存储")
	}
	if raw.Headers == "Authorization: Bearer secret" {
		t.Error("Headers 应以密文存储")
	}

	// 通过 service 读取应还原为明文。
	got, err := svc.Get(ctx, tg.ID)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got.Cookies != "PHPSESSID=secret" {
		t.Errorf("解密 Cookies 不符，得到 %q", got.Cookies)
	}
	if got.Headers != "Authorization: Bearer secret" {
		t.Errorf("解密 Headers 不符，得到 %q", got.Headers)
	}
}

func TestUpdatePreservesSensitiveFields(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	tg := &Target{
		Name:    "test-target",
		URL:     "http://example.com/shell.php",
		Cookies: "PHPSESSID=secret",
	}
	if err := svc.Create(ctx, tg); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	// 仅更新 name，不提供 cookies，应保留原 cookies。
	update := &Target{
		ID:   tg.ID,
		Name: "renamed",
		URL:  "http://example.com/shell.php",
	}
	if err := svc.Update(ctx, update); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	got, err := svc.Get(ctx, tg.ID)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got.Name != "renamed" {
		t.Errorf("期望 name=renamed，得到 %q", got.Name)
	}
	if got.Cookies != "PHPSESSID=secret" {
		t.Errorf("部分更新应保留 Cookies，得到 %q", got.Cookies)
	}
}

func TestUpdateNonExistentReturnsNotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	update := &Target{ID: 9999, Name: "x", URL: "http://example.com/shell.php"}
	if err := svc.Update(ctx, update); err != ErrNotFound {
		t.Errorf("更新不存在的目标应返回 ErrNotFound，得到 %v", err)
	}
}

func TestMaskSensitive(t *testing.T) {
	tg := &Target{Cookies: "secret", Headers: "secret"}
	MaskSensitive(tg)
	if tg.Cookies != "***" || tg.Headers != "***" {
		t.Errorf("掩码后敏感字段应为 ***，得到 cookies=%q headers=%q", tg.Cookies, tg.Headers)
	}
}
