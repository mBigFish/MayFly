package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/webshell-manager/webshell-manager/internal/api/dto"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// TargetHandler 目标相关处理器。
type TargetHandler struct {
	svc *target.Service
}

// NewTargetHandler 创建目标处理器。
func NewTargetHandler(svc *target.Service) *TargetHandler {
	return &TargetHandler{svc: svc}
}

// List 列出目标。
func (h *TargetHandler) List(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	targets, total, err := h.svc.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, "查询目标失败"))
		return
	}

	// 列表场景对敏感字段做掩码，避免明文泄露。
	for i := range targets {
		target.MaskSensitive(&targets[i])
	}

	c.JSON(http.StatusOK, dto.OK(dto.PageData{Items: targets, Total: total}))
}

// Get 获取单个目标。
func (h *TargetHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "目标 ID 非法"))
		return
	}

	t, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, target.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(404, "目标不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(500, "查询目标失败"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(t))
}

// Create 创建目标。
func (h *TargetHandler) Create(c *gin.Context) {
	var req dto.CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}

	t := &target.Target{
		Name:     req.Name,
		URL:      req.URL,
		Type:     req.Type,
		Protocol: req.Protocol,
		Method:   req.Method,
		Headers:  req.Headers,
		Cookies:  req.Cookies,
		Timeout:  req.Timeout,
		Proxy:    req.Proxy,
		Encoding: req.Encoding,
		Remark:   req.Remark,
		GroupID:  req.GroupID,
	}

	if err := h.svc.Create(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	// 重新读取以返回解密后的完整数据。
	created, err := h.svc.Get(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, "创建成功但读取失败"))
		return
	}
	c.JSON(http.StatusCreated, dto.OK(created))
}

// Update 更新目标。
func (h *TargetHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "目标 ID 非法"))
		return
	}

	var req dto.UpdateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}

	t := &target.Target{
		ID:       uint(id),
		Name:     req.Name,
		URL:      req.URL,
		Type:     req.Type,
		Protocol: req.Protocol,
		Method:   req.Method,
		Headers:  req.Headers,
		Cookies:  req.Cookies,
		Timeout:  req.Timeout,
		Proxy:    req.Proxy,
		Encoding: req.Encoding,
		Remark:   req.Remark,
		GroupID:  req.GroupID,
	}

	if err := h.svc.Update(c.Request.Context(), t); err != nil {
		if errors.Is(err, target.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(404, "目标不存在"))
			return
		}
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	updated, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, "更新成功但读取失败"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(updated))
}

// Delete 删除目标。
func (h *TargetHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "目标 ID 非法"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, target.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(404, "目标不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(500, "删除目标失败"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(nil))
}
