package service

import (
	"time"

	"gorm.io/gorm"
	"mayfly/internal/model"
)

// CreateSession 创建会话
func CreateSession(db *gorm.DB, s *model.Session) error {
	return db.Create(s).Error
}

// GetSessionByID 按 ID 查询会话
func GetSessionByID(db *gorm.DB, id uint) (*model.Session, error) {
	var s model.Session
	if err := db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ListSessions 查询会话列表
func ListSessions(db *gorm.DB, userID uint, sessionType string) ([]model.Session, error) {
	var sessions []model.Session
	q := db.Model(&model.Session{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if sessionType != "" {
		q = q.Where("type = ?", sessionType)
	}
	err := q.Order("id DESC").Find(&sessions).Error
	return sessions, err
}

// UpdateSessionActive 更新会话活跃时间
func UpdateSessionActive(db *gorm.DB, id uint) error {
	now := time.Now()
	return db.Exec("UPDATE sessions SET last_active = ?, status = 'active' WHERE id = ?", now, id).Error
}

// CloseSession 关闭会话
func CloseSession(db *gorm.DB, id uint) error {
	return db.Exec("UPDATE sessions SET status = 'closed' WHERE id = ?", id).Error
}

// DeleteSession 删除会话
func DeleteSession(db *gorm.DB, id uint) error {
	return db.Delete(&model.Session{}, id).Error
}
