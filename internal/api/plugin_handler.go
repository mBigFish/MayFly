package api

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/plugin"
	"mayfly/internal/plugin/builtin"
	"mayfly/internal/service"
)

// ensure builtin plugins are registered
func init() {
	builtin.RegisterAll()
}

// ListPlugins 列出所有插件
// GET /api/v1/plugins
func ListPlugins(c *gin.Context) {
	plugins := plugin.GetManager().List()
	list := make([]gin.H, 0, len(plugins))
	for _, p := range plugins {
		list = append(list, gin.H{
			"name":        p.Name(),
			"version":     p.Version(),
			"description": p.Description(),
		})
	}
	OK(c, list)
}

// ExecutePlugin 执行插件
// POST /api/v1/plugins/:name/execute
func ExecutePlugin(c *gin.Context) {
	name := c.Param("name")

	var req struct {
		TargetID uint              `json:"target_id" binding:"required"`
		Params   map[string]string `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	target, err := service.GetTargetByID(database.Get(), req.TargetID)
	if err != nil {
		Fail(c, 404, "目标不存在")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := plugin.GetManager().Execute(ctx, name, target, req.Params)
	if err != nil {
		Fail(c, 500, "插件执行失败: "+err.Error())
		return
	}

	// 记录审计日志
	service.LogAction(database.Get(), c.GetUint("user_id"), c.GetString("username"),
		"plugin_execute", "plugin", 0, "plugin: "+name+", target: "+target.Name, c.ClientIP(), result.Status)

	OK(c, result)
}

// 确保 strconv 被使用
var _ = strconv.Itoa
