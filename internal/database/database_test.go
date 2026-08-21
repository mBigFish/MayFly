package database

import (
	"path/filepath"
	"testing"

	"github.com/webshell-manager/webshell-manager/internal/config"
)

func TestOpenSQLite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	cfg := &config.DatabaseConfig{Driver: "sqlite", Path: path}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open(sqlite) 失败: %v", err)
	}
	if db == nil {
		t.Fatal("db 不应为 nil")
	}
	if err := Migrate(db, nil); err != nil {
		t.Fatalf("Migrate 失败: %v", err)
	}
	if err := Close(); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}
}

func TestOpenUnsupportedDriver(t *testing.T) {
	cfg := &config.DatabaseConfig{Driver: "mysql", Path: "x"}
	if _, err := Open(cfg); err == nil {
		t.Error("不支持的驱动应报错")
	}
}
