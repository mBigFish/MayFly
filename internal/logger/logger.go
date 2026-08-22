package logger

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log *zap.Logger
	su  *zap.SugaredLogger
	mu  sync.RWMutex
)

// Init 初始化日志系统
func Init(level, encoding, output string, maxSize, maxBackups, maxAge int) {
	mu.Lock()
	defer mu.Unlock()

	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 确保日志目录存在
	if output != "" {
		_ = os.MkdirAll(filepath.Dir(output), 0o755)
	}

	// 文件写入器（带轮转）
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   output,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	})

	// 控制台写入器
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 多输出：同时写文件和控制台
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, zapLevel),
		zapcore.NewCore(encoder, consoleWriter, zapLevel),
	)

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	su = log.Sugar()
}

// L 获取原始 Logger
func L() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if log == nil {
		// 默认配置
		Init("info", "console", "data/logs/mayfly.log", 50, 5, 30)
	}
	return log
}

// S 获取 Sugared Logger
func S() *zap.SugaredLogger {
	mu.RLock()
	defer mu.RUnlock()
	if su == nil {
		Init("info", "console", "data/logs/mayfly.log", 50, 5, 30)
	}
	return su
}

// Sync 刷新日志缓冲
func Sync() {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		_ = log.Sync()
	}
}
