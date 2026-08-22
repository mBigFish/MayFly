package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/webshell-manager/webshell-manager/internal/api/dto"
	"github.com/webshell-manager/webshell-manager/internal/auth"
	"github.com/webshell-manager/webshell-manager/internal/session"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// SessionHandler 会话管理处理器。
type SessionHandler struct {
	targetSvc *target.Service
	manager   *session.Manager
}

// NewSessionHandler 创建会话处理器。
func NewSessionHandler(targetSvc *target.Service, manager *session.Manager) *SessionHandler {
	return &SessionHandler{targetSvc: targetSvc, manager: manager}
}

// Create 创建会话（绑定到授权目标）。
func (h *SessionHandler) Create(c *gin.Context) {
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}

	// 读取目标（含解密敏感字段）。
	t, err := h.targetSvc.Get(c.Request.Context(), req.TargetID)
	if err != nil {
		if errors.Is(err, target.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(404, "目标不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(500, "查询目标失败"))
		return
	}

	// 获取当前用户 ID。
	user, ok := c.Get(auth.ContextUserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "未认证"))
		return
	}
	u := user.(*auth.User)

	s, err := h.manager.Create(u.ID, t)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.OK(dto.SessionResponse{
		ID:        s.ID,
		TargetID:  s.TargetID,
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastSeen:  s.LastSeen.Format("2006-01-02T15:04:05Z07:00"),
	}))
}

// List 列出当前用户的会话。
func (h *SessionHandler) List(c *gin.Context) {
	user, ok := c.Get(auth.ContextUserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "未认证"))
		return
	}
	u := user.(*auth.User)

	sessions := h.manager.List(u.ID)
	result := make([]dto.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, dto.SessionResponse{
			ID:        s.ID,
			TargetID:  s.TargetID,
			UserID:    s.UserID,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastSeen:  s.LastSeen.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

// Close 关闭会话。
func (h *SessionHandler) Close(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.Error(400, "会话 ID 不能为空"))
		return
	}
	h.manager.Close(id)
	c.JSON(http.StatusOK, dto.OK(nil))
}
