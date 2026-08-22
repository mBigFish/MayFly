package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/service"
)

// ListTasks 查询任务列表
// GET /api/v1/tasks
func ListTasks(c *gin.Context) {
	keyword := c.Query("keyword")
	tasks, err := service.ListTasks(database.Get(), keyword)
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OK(c, tasks)
}

// GetTask 查询单个任务
// GET /api/v1/tasks/:id
func GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	task, err := service.GetTaskByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "任务不存在")
		return
	}
	OK(c, task)
}

// CreateTask 创建任务
// POST /api/v1/tasks
func CreateTask(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Type    string `json:"type" binding:"required"`
		Payload string `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	task := &model.Task{
		Name:    req.Name,
		Type:    req.Type,
		Payload: req.Payload,
		Status:  "pending",
		UserID:  c.GetUint("user_id"),
		Username: c.GetString("username"),
	}

	if err := service.CreateTask(database.Get(), task); err != nil {
		Fail(c, 500, "创建失败: "+err.Error())
		return
	}

	// 提交到工作池异步执行
	service.GetTaskWorker().Submit(task.ID)

	OKMsg(c, "创建成功，已提交执行", task)
}

// CancelTask 取消任务
// POST /api/v1/tasks/:id/cancel
func CancelTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.UpdateTaskStatus(database.Get(), uint(id), "cancelled"); err != nil {
		Fail(c, 500, "取消失败: "+err.Error())
		return
	}
	OKMsg(c, "已取消", nil)
}

// DeleteTask 删除任务
// DELETE /api/v1/tasks/:id
func DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteTask(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "删除失败: "+err.Error())
		return
	}
	OKMsg(c, "已删除", nil)
}
