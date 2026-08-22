package api

import (
	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
)

// GetDashboard 仪表盘数据
// GET /api/v1/dashboard
func GetDashboard(c *gin.Context) {
	db := database.Get()

	var targetTotal, targetOnline, sessionActive, listenerRunning, taskTotal, auditTotal, userTotal int64

	db.Model(&model.Target{}).Count(&targetTotal)
	db.Model(&model.Target{}).Where("status = ?", "online").Count(&targetOnline)
	db.Model(&model.Session{}).Where("status = ?", "active").Count(&sessionActive)
	db.Model(&model.Listener{}).Where("status = ?", "running").Count(&listenerRunning)
	db.Model(&model.Task{}).Count(&taskTotal)
	db.Model(&model.AuditLog{}).Count(&auditTotal)
	db.Model(&model.User{}).Count(&userTotal)

	OK(c, gin.H{
		"targets":         targetTotal,
		"targets_online":  targetOnline,
		"sessions_active": sessionActive,
		"listeners":       listenerRunning,
		"tasks":           taskTotal,
		"audit_logs":      auditTotal,
		"users":           userTotal,
	})
}
