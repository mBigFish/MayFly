package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/service"
)

// ListSessions 查询会话列表
// GET /api/v1/sessions?type=xxx
func ListSessions(c *gin.Context) {
	sessionType := c.Query("type")
	userID := c.GetUint("user_id")
	sessions, err := service.ListSessions(database.Get(), userID, sessionType)
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OK(c, sessions)
}

// GetSession 查询单个会话
// GET /api/v1/sessions/:id
func GetSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	session, err := service.GetSessionByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "会话不存在")
		return
	}
	OK(c, session)
}

// CloseSession 关闭会话
// POST /api/v1/sessions/:id/close
func CloseSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.CloseSession(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "关闭失败: "+err.Error())
		return
	}
	OKMsg(c, "已关闭", nil)
}

// DeleteSession 删除会话
// DELETE /api/v1/sessions/:id
func DeleteSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteSession(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "删除失败: "+err.Error())
		return
	}
	OKMsg(c, "已删除", nil)
}
