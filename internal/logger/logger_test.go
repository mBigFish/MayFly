package logger

import (
	"path/filepath"
	"testing"
)

func TestInitValidLevel(t *testing.T) {
	if err := Init("info", ""); err != nil {
		t.Fatalf("Init(info) 不应报错: %v", err)
	}
	if Logger == nil {
		t.Fatal("Logger 不应为 nil")
	}
}

func TestInitInvalidLevel(t *testing.T) {
	if err := Init("notalevel", ""); err == nil {
		t.Error("无效日志级别应报错")
	}
}

func TestInitWithFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "app.log")
	if err := Init("debug", path); err != nil {
		t.Fatalf("Init(debug, file) 不应报错: %v", err)
	}
	Logger.Info("test log message")
	Sync()
}
