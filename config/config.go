package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"gopkg.in/yaml.v3"
)

// resolvePath 优先使用可执行文件所在目录下的资源路径，找不到时回退到当前工作目录。
func resolvePath(sub string) string {
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(filepath.Join(filepath.Dir(exe), sub)); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return sub
}

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
	loadYAML(resolvePath("config/config.yaml"))

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

// Save 将当前配置写回 config.yaml 文件，使运行时修改持久化到配置文件。
func Save() error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}
	content := "# Mayfly WebShell 管理器配置\n"
	content += "# 配置优先级：环境变量 > 本文件 > 默认值\n"
	content += "# 即：如果设置了对应的环境变量（如 MAYFLY_PORT），会覆盖本文件的值。\n\n"
	content += fmt.Sprintf("# 服务器监听端口\nserver_port: %q\n\n", cfg.ServerPort)
	content += "# 认证配置\n"
	content += fmt.Sprintf("username: %q\n", cfg.Username)
	content += fmt.Sprintf("password: %q\n", cfg.Password)
	content += "# JWT 签名密钥（生产环境务必修改为随机长字符串）\n"
	content += fmt.Sprintf("jwt_secret: %q\n\n", cfg.JWTSecret)
	content += "# 默认终端 shell（Windows 为 powershell.exe，Linux/macOS 为 bash）\n"
	content += fmt.Sprintf("shell: %q\n\n", cfg.Shell)
	content += "# 会话超时（分钟）\n"
	content += fmt.Sprintf("session_timeout: %d\n", cfg.SessionTimeout)
	path := resolvePath("config/config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// defaultShell 根据操作系统返回默认 shell
func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "bash"
}
