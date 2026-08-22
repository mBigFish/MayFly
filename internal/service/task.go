package service

import (
	"gorm.io/gorm"
	"mayfly/internal/model"
)

// ListTasks 查询任务列表
func ListTasks(db *gorm.DB, keyword string) ([]model.Task, error) {
	var tasks []model.Task
	q := db.Model(&model.Task{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	err := q.Order("id DESC").Find(&tasks).Error
	return tasks, err
}

// GetTaskByID 按 ID 查询任务
func GetTaskByID(db *gorm.DB, id uint) (*model.Task, error) {
	var task model.Task
	if err := db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTask 创建任务
func CreateTask(db *gorm.DB, t *model.Task) error {
	return db.Create(t).Error
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(db *gorm.DB, id uint, status string) error {
	return db.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, id).Error
}

// UpdateTaskProgress 更新任务进度
func UpdateTaskProgress(db *gorm.DB, id uint, done, total int) error {
	return db.Exec("UPDATE tasks SET done = ?, total = ? WHERE id = ?", done, total, id).Error
}

// DeleteTask 删除任务
func DeleteTask(db *gorm.DB, id uint) error {
	return db.Delete(&model.Task{}, id).Error
}
