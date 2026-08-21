// Package logger 基于 zap 提供统一的结构化日志能力。
package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是应用使用的全局日志入口。
var Logger *zap.Logger

// Init 根据配置初始化全局 Logger。
// 日志同时输出到控制台与指定文件（若配置了路径）。
func Init(level string, path string) error {
	lvl, err := parseLevel(level)
	if err != nil {
		return err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	cores := []zapcore.Core{
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			lvl,
		),
	}

	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(file),
			lvl,
		))
	}

	core := zapcore.NewTee(cores...)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// parseLevel 将字符串级别解析为 zapcore.Level。
func parseLevel(level string) (zapcore.Level, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return 0, fmt.Errorf("无效的日志级别 %q: %w", level, err)
	}
	return lvl, nil
}

// Sync 刷新并关闭日志缓冲区，应在程序退出前调用。
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
