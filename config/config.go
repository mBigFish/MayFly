package config

import (
	"os"
	"runtime"
	"strconv"
)

// Config 全局配置
type Config struct {
	// 服务器配置
	ServerPort string `json:"server_port"`
	// 认证配置
	Username string `json:"username"`
	Password string `json:"password"`
	JWTSecret string `json:"jwt_secret"`
	// 终端配置
	Shell string `json:"shell"`
	// 会话超时（分钟）
	SessionTimeout int `json:"session_timeout"`
}

var cfg *Config

// Init 初始化配置，从环境变量读取，提供默认值
func Init() {
	cfg = &Config{
		ServerPort:     getEnv("MAYFLY_PORT", "8080"),
		Username:      getEnv("MAYFLY_USER", "admin"),
		Password:      getEnv("MAYFLY_PASS", "mayfly123"),
		JWTSecret:     getEnv("MAYFLY_JWT_SECRET", "mayfly-secret-key-change-in-production"),
		Shell:         getEnv("MAYFLY_SHELL", defaultShell()),
		SessionTimeout: getEnvInt("MAYFLY_SESSION_TIMEOUT", 30),
	}
}

// Get 获取配置实例
func Get() *Config {
	if cfg == nil {
		Init()
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

// defaultShell 根据操作系统返回默认 shell
func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "bash"
}
