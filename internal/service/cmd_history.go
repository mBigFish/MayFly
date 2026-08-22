package service

import (
	"gorm.io/gorm"
	"mayfly/internal/model"
)

// ListCmdHistory 查询命令历史
func ListCmdHistory(db *gorm.DB, targetID uint, page, perPage int) ([]model.CmdHistory, int64, error) {
	var history []model.CmdHistory
	var total int64
	q := db.Model(&model.CmdHistory{}).Where("target_id = ?", targetID)
	q.Count(&total)
	offset := (page - 1) * perPage
	err := q.Order("id DESC").Offset(offset).Limit(perPage).Find(&history).Error
	return history, total, err
}

// CreateCmdHistory 创建命令历史
func CreateCmdHistory(db *gorm.DB, h *model.CmdHistory) error {
	return db.Create(h).Error
}

// DeleteCmdHistory 删除命令历史
func DeleteCmdHistory(db *gorm.DB, targetID uint) error {
	return db.Where("target_id = ?", targetID).Delete(&model.CmdHistory{}).Error
}

// ClearAllCmdHistory 清空所有命令历史
func ClearAllCmdHistory(db *gorm.DB) error {
	return db.Exec("DELETE FROM cmd_history").Error
}
