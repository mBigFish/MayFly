package service

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"mayfly/internal/model"
)

// LogRequest 记录请求日志
func LogRequest(db *gorm.DB, log *model.RequestLog) {
	if log == nil {
		return
	}
	// 限制 response 长度
	if len(log.Response) > 65536 {
		log.Response = log.Response[:65536] + "...(truncated)"
	}
	if len(log.Request) > 65536 {
		log.Request = log.Request[:65536] + "...(truncated)"
	}
	db.Create(log)
}

// ListRequestLogs 列出请求日志
func ListRequestLogs(db *gorm.DB, targetID uint, keyword string, page, pageSize int) ([]*model.RequestLog, int64, error) {
	var logs []*model.RequestLog
	var total int64

	q := db.Model(&model.RequestLog{})
	if targetID > 0 {
		q = q.Where("target_id = ?", targetID)
	}
	if keyword != "" {
		q = q.Where("operation LIKE ? OR target_name LIKE ? OR username LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	q.Count(&total)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	return logs, total, err
}

// GetRequestLog 获取请求日志详情
func GetRequestLog(db *gorm.DB, id uint) (*model.RequestLog, error) {
	var log model.RequestLog
	err := db.First(&log, id).Error
	return &log, err
}

// formatDuration 格式化耗时
func formatDuration(d time.Duration) int64 {
	return d.Milliseconds()
}

// ensure json is used
var _ = json.Marshal
