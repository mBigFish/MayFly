// Package database 负责数据库连接的初始化与管理。
// 第一版使用 SQLite，通过 GORM 预留 PostgreSQL 扩展能力。
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/webshell-manager/webshell-manager/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局数据库连接实例。
var DB *gorm.DB

// Open 根据配置建立数据库连接。
func Open(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite":
		// 确保 SQLite 文件所在目录存在。
		if dir := filepath.Dir(cfg.Path); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("创建数据库目录失败: %w", err)
			}
		}
		dialector = sqlite.Open(cfg.Path)
	case "postgres":
		dialector = postgres.Open(cfg.Path)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %q", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层 sql.DB 以配置连接池参数。
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	DB = db
	return db, nil
}

// Close 关闭数据库连接。
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
