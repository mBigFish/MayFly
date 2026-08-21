// Package migrations 定义应用的具体数据库迁移与数据初始化。
// 独立成包以避免 database 包与业务包之间的循环依赖。
package migrations

import (
	"fmt"

	"github.com/webshell-manager/webshell-manager/internal/auth"
	"github.com/webshell-manager/webshell-manager/internal/database"
	"github.com/webshell-manager/webshell-manager/internal/target"
	"gorm.io/gorm"
)

// All 返回按顺序执行的迁移列表。
func All() []database.Migration {
	return []database.Migration{
		{Name: "create_core_tables", Run: createCoreTables},
		{Name: "seed_roles_and_permissions", Run: seedRolesAndPermissions},
	}
}

// createCoreTables 创建用户、角色、权限、目标等核心表。
func createCoreTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&auth.User{},
		&auth.Role{},
		&auth.Permission{},
		&target.Target{},
		&target.TargetGroup{},
	)
}

// seedRolesAndPermissions 初始化权限与内置角色。
func seedRolesAndPermissions(db *gorm.DB) error {
	// 写入全部权限码。
	for _, code := range auth.AllPermissions() {
		var perm auth.Permission
		err := db.Where("code = ?", code).First(&perm).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&auth.Permission{Code: code}).Error; err != nil {
				return fmt.Errorf("创建权限 %q 失败: %w", code, err)
			}
		} else if err != nil {
			return err
		}
	}

	// 写入内置角色并关联权限。
	for roleName, permCodes := range auth.RolePermissionMap() {
		var role auth.Role
		err := db.Where("name = ?", roleName).First(&role).Error
		if err == gorm.ErrRecordNotFound {
			role = auth.Role{Name: roleName}
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("创建角色 %q 失败: %w", roleName, err)
			}
		} else if err != nil {
			return err
		}

		for _, code := range permCodes {
			var perm auth.Permission
			if err := db.Where("code = ?", code).First(&perm).Error; err != nil {
				return fmt.Errorf("查询权限 %q 失败: %w", code, err)
			}
			if err := db.Model(&role).Association("Permissions").Append(&perm); err != nil {
				return fmt.Errorf("关联角色 %q 权限 %q 失败: %w", roleName, code, err)
			}
		}
	}

	return nil
}
