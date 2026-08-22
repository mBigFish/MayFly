package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	Terminal TerminalConfig `yaml:"terminal"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type AuthConfig struct {
	JWTSecret    string `yaml:"jwt_secret"`
	JWTExpire    int    `yaml:"jwt_expire"`
	InitUsername string `yaml:"init_username"`
	InitPassword string `yaml:"init_password"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	Encoding   string `yaml:"encoding"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type TerminalConfig struct {
	Shell          string `yaml:"shell"`
	SessionTimeout int    `yaml:"session_timeout"`
}

var (
	cfg     *Config
	cfgOnce sync.Once
	cfgMu   sync.RWMutex
)

// Load 从 YAML 文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	c.applyDefaults()
	return &c, nil
}

// applyDefaults 填充默认值
func (c *Config) applyDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.Auth.JWTSecret == "" {
		c.Auth.JWTSecret = "mayfly-secret-key-change-in-production"
	}
	if c.Auth.JWTExpire <= 0 {
		c.Auth.JWTExpire = 30
	}
	if c.Auth.InitUsername == "" {
		c.Auth.InitUsername = "admin"
	}
	if c.Auth.InitPassword == "" {
		c.Auth.InitPassword = "mayfly123"
	}
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite"
	}
	if c.Database.DSN == "" {
		c.Database.DSN = "data/mayfly.db"
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Encoding == "" {
		c.Log.Encoding = "console"
	}
	if c.Log.Output == "" {
		c.Log.Output = "data/logs/mayfly.log"
	}
	if c.Log.MaxSize <= 0 {
		c.Log.MaxSize = 50
	}
	if c.Log.MaxBackups <= 0 {
		c.Log.MaxBackups = 5
	}
	if c.Log.MaxAge <= 0 {
		c.Log.MaxAge = 30
	}
	if c.Terminal.Shell == "" {
		c.Terminal.Shell = defaultShell()
	}
	if c.Terminal.SessionTimeout <= 0 {
		c.Terminal.SessionTimeout = 30
	}
}

// defaultShell 根据操作系统返回默认 shell
func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "bash"
}

// Init 初始化全局配置（从指定路径加载）
func Init(path string) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	c, err := Load(path)
	if err != nil {
		return err
	}
	cfg = c
	return nil
}

// Get 获取全局配置实例
func Get() *Config {
	cfgOnce.Do(func() {
		if cfg == nil {
			// 尝试默认路径
			path := "configs/config.yaml"
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// 尝试可执行文件目录
				if exe, e2 := os.Executable(); e2 == nil {
					p2 := filepath.Join(filepath.Dir(exe), "configs", "config.yaml")
					if _, e3 := os.Stat(p2); e3 == nil {
						path = p2
					}
				}
			}
			c, err := Load(path)
			if err != nil {
				// 使用默认配置
				c = &Config{}
				c.applyDefaults()
			}
			cfg = c
		}
	})
	return cfg
}

// Set 设置全局配置（测试用）
func Set(c *Config) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfg = c
}
