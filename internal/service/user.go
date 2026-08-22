package service

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mayfly/internal/model"
)

// CreateUser 创建用户
func CreateUser(db *gorm.DB, username, password, nickname, email, role string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: username,
		Password: string(hashed),
		Nickname: nickname,
		Email:    email,
		Role:     role,
		Status:   1,
	}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUsername 按用户名查询用户
func GetUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 按 ID 查询用户
func GetUserByID(db *gorm.DB, id uint) (*model.User, error) {
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// VerifyPassword 验证密码
func VerifyPassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// UpdateLastLogin 更新最后登录时间
func UpdateLastLogin(db *gorm.DB, userID uint) error {
	now := time.Now()
	return db.Exec("UPDATE users SET last_login_at = ?, updated_at = ? WHERE id = ?", now, now, userID).Error
}

// ChangePassword 修改密码
func ChangePassword(db *gorm.DB, userID uint, oldPassword, newPassword string) error {
	user, err := GetUserByID(db, userID)
	if err != nil {
		return err
	}
	if err := VerifyPassword(user.Password, oldPassword); err != nil {
		return errors.New("旧密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Exec("UPDATE users SET password = ?, updated_at = ? WHERE id = ?", string(hashed), time.Now(), userID).Error
}

// InitAdminIfEmpty 如果 users 表为空，创建初始管理员
func InitAdminIfEmpty(db *gorm.DB, username, password string) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}
	_, err := CreateUser(db, username, password, "Administrator", "", "admin")
	return err
}
