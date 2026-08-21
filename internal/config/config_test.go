package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("默认 host 应为 0.0.0.0，得到 %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("默认 port 应为 8080，得到 %d", cfg.Server.Port)
	}
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("默认 driver 应为 sqlite，得到 %q", cfg.Database.Driver)
	}
	if cfg.Security.SessionTimeout != 3600 {
		t.Errorf("默认 session_timeout 应为 3600，得到 %d", cfg.Security.SessionTimeout)
	}
}

func TestLoadEmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") 不应报错: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("空路径应返回默认配置，得到 port=%d", cfg.Server.Port)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "server:\n  host: \"127.0.0.1\"\n  port: 9090\nlogging:\n  level: debug\n  path: \"/tmp/x.log\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("期望 port=9090，得到 %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("期望 level=debug，得到 %q", cfg.Logging.Level)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load("/nonexistent/path/config.yaml"); err == nil {
		t.Error("加载不存在的文件应报错")
	}
}

func TestAddr(t *testing.T) {
	cfg := Default()
	if cfg.Addr() != "0.0.0.0:8080" {
		t.Errorf("Addr 期望 0.0.0.0:8080，得到 %q", cfg.Addr())
	}
}
