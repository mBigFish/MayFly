package config

import (
	"os"
	"runtime"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	// 服务器配置
	ServerPort string `yaml:"server_port" json:"server_port"`
	// 认证配置
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	JWTSecret string `yaml:"jwt_secret" json:"jwt_secret"`
	// 终端配置
	Shell string `yaml:"shell" json:"shell"`
	// 会话超时（分钟）
	SessionTimeout int `yaml:"session_timeout" json:"session_timeout"`
}

var cfg *Config

// Init 初始化配置。
// 优先级：config.yaml 文件 > 环境变量 > 默认值。
func Init() {
	// 1. 默认值
	cfg = &Config{
		ServerPort:     "8080",
		Username:       "admin",
		Password:       "mayfly123",
		JWTSecret:      "mayfly-secret-key-change-in-production",
		Shell:          defaultShell(),
		SessionTimeout: 30,
	}

	// 2. 加载 config.yaml（若存在），覆盖默认值
	loadYAML("config/config.yaml")

	// 3. 环境变量覆盖（优先级最高）
	if v := os.Getenv("MAYFLY_PORT"); v != "" {
		cfg.ServerPort = v
	}
	if v := os.Getenv("MAYFLY_USER"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("MAYFLY_PASS"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("MAYFLY_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("MAYFLY_SHELL"); v != "" {
		cfg.Shell = v
	}
	if v := os.Getenv("MAYFLY_SESSION_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SessionTimeout = i
		}
	}
}

// loadYAML 从指定路径加载配置文件，覆盖 cfg 中非零值的字段。
// 文件不存在时静默忽略（回退到环境变量/默认值）。
func loadYAML(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return
	}
	// 仅覆盖文件里显式提供的字段（非零值），其余保留默认值
	if fileCfg.ServerPort != "" {
		cfg.ServerPort = fileCfg.ServerPort
	}
	if fileCfg.Username != "" {
		cfg.Username = fileCfg.Username
	}
	if fileCfg.Password != "" {
		cfg.Password = fileCfg.Password
	}
	if fileCfg.JWTSecret != "" {
		cfg.JWTSecret = fileCfg.JWTSecret
	}
	if fileCfg.Shell != "" {
		cfg.Shell = fileCfg.Shell
	}
	if fileCfg.SessionTimeout != 0 {
		cfg.SessionTimeout = fileCfg.SessionTimeout
	}
}

// Get 获取配置实例
func Get() *Config {
	if cfg == nil {
		Init()
	}
	return cfg
}

// defaultShell 根据操作系统返回默认 shell
func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "bash"
}
