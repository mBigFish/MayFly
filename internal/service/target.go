package service

import (
	"gorm.io/gorm"
	"mayfly/internal/model"
)

// ListTargets 查询目标列表
func ListTargets(db *gorm.DB, keyword string) ([]model.Target, error) {
	var targets []model.Target
	q := db.Model(&model.Target{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR url LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := q.Order("id DESC").Find(&targets).Error
	return targets, err
}

// GetTargetByID 按 ID 查询目标
func GetTargetByID(db *gorm.DB, id uint) (*model.Target, error) {
	var target model.Target
	if err := db.First(&target, id).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

// CreateTarget 创建目标
func CreateTarget(db *gorm.DB, t *model.Target) error {
	return db.Create(t).Error
}

// UpdateTarget 更新目标
func UpdateTarget(db *gorm.DB, t *model.Target) error {
	return db.Save(t).Error
}

// DeleteTarget 删除目标
func DeleteTarget(db *gorm.DB, id uint) error {
	return db.Delete(&model.Target{}, id).Error
}

// UpdateTargetStatus 更新目标状态
func UpdateTargetStatus(db *gorm.DB, id uint, status string) error {
	return db.Model(&model.Target{}).Where("id = ?", id).Update("status", status).Error
}
