// Package config 负责加载与管理应用配置。
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 是应用根配置结构。
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Security SecurityConfig `yaml:"security"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig HTTP 服务配置。
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig 数据库配置。
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}

// SecurityConfig 安全相关配置。
type SecurityConfig struct {
	SessionTimeout int    `yaml:"session_timeout"`
	JWTSecret      string `yaml:"jwt_secret"`
	EncryptionKey  string `yaml:"encryption_key"`
}

// LoggingConfig 日志配置。
type LoggingConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

// Load 从指定路径加载配置文件。
// 若 path 为空则使用默认值。
func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil
}

// Default 返回带默认值的配置。
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			Path:   "./data/app.db",
		},
		Security: SecurityConfig{
			SessionTimeout: 3600,
			JWTSecret:      "change-me-to-a-random-secret",
			EncryptionKey:  "0123456789abcdef0123456789abcdef",
		},
		Logging: LoggingConfig{
			Level: "info",
			Path:  "./logs/app.log",
		},
	}
}

// Addr 返回服务监听地址。
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
