// Package database 迁移框架。具体迁移在 internal/migrations 中定义，
// 以避免 database 包与业务包之间的循环依赖。
package database

import (
	"fmt"

	"gorm.io/gorm"
)

// Migration 表示一条可执行的结构迁移。
type Migration struct {
	Name string
	Run  func(*gorm.DB) error
}

// Migrate 依次执行给定的迁移列表。
func Migrate(db *gorm.DB, migrations []Migration) error {
	for _, m := range migrations {
		if err := m.Run(db); err != nil {
			return fmt.Errorf("执行迁移 %q 失败: %w", m.Name, err)
		}
	}
	return nil
}
