package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/webshell-manager/webshell-manager/internal/api/dto"
	"github.com/webshell-manager/webshell-manager/internal/protocol"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// TargetCheckHandler 目标探活处理器。
type TargetCheckHandler struct {
	targetSvc *target.Service
	registry  *protocol.Registry
}

// NewTargetCheckHandler 创建目标探活处理器。
func NewTargetCheckHandler(targetSvc *target.Service, registry *protocol.Registry) *TargetCheckHandler {
	return &TargetCheckHandler{targetSvc: targetSvc, registry: registry}
}

// Check 探活指定目标。
// 架构：API → Service(取目标) → Protocol.Check → Transport → 授权目标。
func (h *TargetCheckHandler) Check(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "目标 ID 非法"))
		return
	}

	// 读取目标（含解密后的敏感字段）。
	t, err := h.targetSvc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, target.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(404, "目标不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(500, "查询目标失败"))
		return
	}

	// 根据目标的 protocol 字段选择协议适配器。
	p, err := h.registry.Get(t.Protocol)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	// 执行探活。
	if err := p.Check(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusOK, dto.OK(dto.CheckResponse{
			TargetID: t.ID,
			Success:  false,
			Message:  err.Error(),
		}))
		return
	}

	c.JSON(http.StatusOK, dto.OK(dto.CheckResponse{
		TargetID: t.ID,
		Success:  true,
		Message:  "探活成功",
	}))
}
