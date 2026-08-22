package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/service"
)

// ListListeners 查询监听器列表
// GET /api/v1/listeners
func ListListeners(c *gin.Context) {
	listeners, err := service.ListListeners(database.Get())
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OK(c, listeners)
}

// CreateListener 创建监听器
// POST /api/v1/listeners
func CreateListener(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" binding:"required"`
		Protocol string `json:"protocol"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	l := &model.Listener{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		Protocol: req.Protocol,
		Status:   "stopped",
		UserID:   c.GetUint("user_id"),
	}
	if l.Protocol == "" {
		l.Protocol = "tcp"
	}

	if err := service.CreateListener(database.Get(), l); err != nil {
		Fail(c, 500, "创建失败: "+err.Error())
		return
	}
	OKMsg(c, "创建成功", l)
}

// StartListener 启动监听器
// POST /api/v1/listeners/:id/start
func StartListener(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	mgr := service.GetListenerManager()
	if err := mgr.StartListener(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "启动失败: "+err.Error())
		return
	}
	OKMsg(c, "已启动", nil)
}

// StopListener 停止监听器
// POST /api/v1/listeners/:id/stop
func StopListener(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	mgr := service.GetListenerManager()
	if err := mgr.StopListener(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "停止失败: "+err.Error())
		return
	}
	OKMsg(c, "已停止", nil)
}

// DeleteListener 删除监听器
// DELETE /api/v1/listeners/:id
func DeleteListener(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteListener(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "删除失败: "+err.Error())
		return
	}
	OKMsg(c, "已删除", nil)
}
