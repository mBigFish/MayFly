package auth

import (
	"context"
	"fmt"

	"github.com/webshell-manager/webshell-manager/internal/database"
	"gorm.io/gorm"
)

// EnsureAdminUser 确保存在初始 admin 用户。
// 若数据库中不存在 username 对应用户则创建，密码使用 bcrypt 哈希后保存。
func EnsureAdminUser(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("admin 用户名或密码不能为空")
	}

	var user User
	err := database.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == nil {
		return nil // 已存在，不覆盖。
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	newUser := User{
		Username:     username,
		PasswordHash: hash,
		Enabled:      true,
	}
	if err := database.DB.WithContext(ctx).Create(&newUser).Error; err != nil {
		return err
	}

	// 绑定 admin 角色。
	var adminRole Role
	if err := database.DB.WithContext(ctx).Where("name = ?", RoleAdmin).First(&adminRole).Error; err != nil {
		return fmt.Errorf("查询 admin 角色失败: %w", err)
	}
	if err := database.DB.WithContext(ctx).Model(&newUser).Association("Roles").Append(&adminRole); err != nil {
		return fmt.Errorf("绑定 admin 角色失败: %w", err)
	}

	return nil
}
