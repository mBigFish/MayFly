package target

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound 目标不存在。
var ErrNotFound = errors.New("目标不存在")

// Repository 目标仓储，封装对 targets 表的访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建目标。
func (r *Repository) Create(ctx context.Context, t *Target) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// GetByID 按 ID 查询目标。
func (r *Repository) GetByID(ctx context.Context, id uint) (*Target, error) {
	var t Target
	err := r.db.WithContext(ctx).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// List 分页查询目标。
func (r *Repository) List(ctx context.Context, offset, limit int) ([]Target, int64, error) {
	var targets []Target
	var total int64

	if err := r.db.WithContext(ctx).Model(&Target{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&targets).Error
	if err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

// Exists 判断目标是否存在。
func (r *Repository) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Target{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update 更新目标的非零字段，避免覆盖未提供的内容。
func (r *Repository) Update(ctx context.Context, t *Target) error {
	// 先确认记录存在，避免误插入。
	exists, err := r.Exists(ctx, t.ID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	return r.db.WithContext(ctx).Model(&Target{}).
		Where("id = ?", t.ID).
		Updates(map[string]interface{}{
			"name":     t.Name,
			"url":      t.URL,
			"type":     t.Type,
			"protocol": t.Protocol,
			"method":   t.Method,
			"headers":  t.Headers,
			"cookies":  t.Cookies,
			"timeout":  t.Timeout,
			"proxy":    t.Proxy,
			"encoding": t.Encoding,
			"remark":   t.Remark,
			"group_id": t.GroupID,
		}).Error
}

// Delete 删除目标。
func (r *Repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&Target{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
