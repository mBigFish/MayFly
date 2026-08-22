package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/service"
)

// ListTargets 查询目标列表
// GET /api/v1/targets?keyword=xxx
func ListTargets(c *gin.Context) {
	keyword := c.Query("keyword")
	targets, err := service.ListTargets(database.Get(), keyword)
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OK(c, targets)
}

// GetTarget 查询单个目标
// GET /api/v1/targets/:id
func GetTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}
	OK(c, target)
}

// CreateTarget 创建目标
// POST /api/v1/targets
func CreateTarget(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		URL      string `json:"url" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Password string `json:"password"`
		Encoding string `json:"encoding"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target := &model.Target{
		Name:     req.Name,
		URL:      req.URL,
		Type:     req.Type,
		Password: req.Password,
		Encoding: req.Encoding,
		Remark:   req.Remark,
		Status:   "unknown",
	}
	if target.Encoding == "" {
		target.Encoding = "auto"
	}

	if err := service.CreateTarget(database.Get(), target); err != nil {
		Fail(c, 500, "创建失败: "+err.Error())
		return
	}
	OKMsg(c, "创建成功", target)
}

// UpdateTarget 更新目标
// PUT /api/v1/targets/:id
func UpdateTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	var req struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Type     string `json:"type"`
		Password string `json:"password"`
		Encoding string `json:"encoding"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	if req.Name != "" {
		target.Name = req.Name
	}
	if req.URL != "" {
		target.URL = req.URL
	}
	if req.Type != "" {
		target.Type = req.Type
	}
	if req.Password != "" {
		target.Password = req.Password
	}
	if req.Encoding != "" {
		target.Encoding = req.Encoding
	}
	if req.Remark != "" {
		target.Remark = req.Remark
	}

	if err := service.UpdateTarget(database.Get(), target); err != nil {
		Fail(c, 500, "更新失败")
		return
	}
	OKMsg(c, "更新成功", target)
}

// DeleteTarget 删除目标
// DELETE /api/v1/targets/:id
func DeleteTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteTarget(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "删除失败")
		return
	}
	OKMsg(c, "删除成功", nil)
}
