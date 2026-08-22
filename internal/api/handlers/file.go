package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/webshell-manager/webshell-manager/internal/api/dto"
	"github.com/webshell-manager/webshell-manager/internal/file"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// FileHandler 文件管理处理器。
type FileHandler struct {
	targetSvc *target.Service
	fileSvc   *file.Service
}

// NewFileHandler 创建文件管理处理器。
func NewFileHandler(targetSvc *target.Service, fileSvc *file.Service) *FileHandler {
	return &FileHandler{targetSvc: targetSvc, fileSvc: fileSvc}
}

// getTarget 解析目标 ID 并读取（解密敏感字段）。
func (h *FileHandler) getTarget(c *gin.Context) (*target.Target, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "目标 ID 非法"))
		return nil, false
	}
	t, err := h.targetSvc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, target.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(404, "目标不存在"))
			return nil, false
		}
		c.JSON(http.StatusInternalServerError, dto.Error(500, "查询目标失败"))
		return nil, false
	}
	return t, true
}

// List 列出目录。path 通过 query 参数传递。
func (h *FileHandler) List(c *gin.Context) {
	t, ok := h.getTarget(c)
	if !ok {
		return
	}
	path := c.Query("path")
	res, err := h.fileSvc.List(c.Request.Context(), t, path)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

// Read 读取文件。
func (h *FileHandler) Read(c *gin.Context) {
	t, ok := h.getTarget(c)
	if !ok {
		return
	}
	var req dto.FileReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}
	res, err := h.fileSvc.Read(c.Request.Context(), t, req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

// Write 写入文件。
func (h *FileHandler) Write(c *gin.Context) {
	t, ok := h.getTarget(c)
	if !ok {
		return
	}
	var req dto.FileWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}
	res, err := h.fileSvc.Write(c.Request.Context(), t, req.Path, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

// Rename 重命名。
func (h *FileHandler) Rename(c *gin.Context) {
	t, ok := h.getTarget(c)
	if !ok {
		return
	}
	var req dto.FileRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}
	res, err := h.fileSvc.Rename(c.Request.Context(), t, req.From, req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

// Mkdir 创建目录。
func (h *FileHandler) Mkdir(c *gin.Context) {
	t, ok := h.getTarget(c)
	if !ok {
		return
	}
	var req dto.FileMkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}
	res, err := h.fileSvc.Mkdir(c.Request.Context(), t, req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

// Delete 删除文件/目录。
func (h *FileHandler) Delete(c *gin.Context) {
	t, ok := h.getTarget(c)
	if !ok {
		return
	}
	var req dto.FileDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请求参数非法"))
		return
	}
	res, err := h.fileSvc.Delete(c.Request.Context(), t, req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}
