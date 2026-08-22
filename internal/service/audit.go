package service

import (
	"gorm.io/gorm"
	"mayfly/internal/model"
)

// CreateAuditLog 创建审计日志
func CreateAuditLog(db *gorm.DB, log *model.AuditLog) error {
	return db.Create(log).Error
}

// ListAuditLogs 查询审计日志
func ListAuditLogs(db *gorm.DB, keyword string, page, perPage int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	q := db.Model(&model.AuditLog{})
	if keyword != "" {
		q = q.Where("username LIKE ? OR action LIKE ? OR resource LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Count(&total)
	offset := (page - 1) * perPage
	err := q.Order("id DESC").Offset(offset).Limit(perPage).Find(&logs).Error
	return logs, total, err
}

// LogAction 记录操作审计日志（便捷方法）
func LogAction(db *gorm.DB, userID uint, username, action, resource string, resourceID uint, detail, ip, status string) {
	_ = CreateAuditLog(db, &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		Status:     status,
	})
}
