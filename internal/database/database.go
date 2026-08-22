package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	applog "mayfly/internal/logger"
)

var (
	db  *gorm.DB
	mu  sync.RWMutex
)

// Init 初始化数据库连接
func Init(driver, dsn string) error {
	mu.Lock()
	defer mu.Unlock()

	// 确保 SQLite 文件目录存在
	if driver == "sqlite" {
		dir := filepath.Dir(dsn)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0o755)
		}
	}

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	d, err := gorm.Open(sqlite.Open(dsn), gormCfg)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 启用 WAL 模式提升并发性能
	if driver == "sqlite" {
		_ = d.Exec("PRAGMA journal_mode=WAL;").Error
	}
	db = d

	applog.S().Info("数据库连接成功", "driver", driver, "dsn", dsn)
	return nil
}

// Get 获取数据库实例
func Get() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	if db == nil {
		// 自动初始化
		_ = Init("sqlite", "data/mayfly.db")
	}
	return db
}

// AutoMigrate 自动迁移模型
func AutoMigrate(models ...interface{}) error {
	return Get().AutoMigrate(models...)
}
