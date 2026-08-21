// Package auth 定义用户、角色、权限实体及认证授权逻辑。
package auth

import "time"

// User 系统用户。
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Enabled      bool      `gorm:"default:true" json:"enabled"`
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (User) TableName() string { return "users" }

// Role 角色。
type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// TableName 指定表名。
func (Role) TableName() string { return "roles" }

// Permission 权限。
type Permission struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Code string `gorm:"size:128;uniqueIndex;not null" json:"code"`
}

// TableName 指定表名。
func (Permission) TableName() string { return "permissions" }
