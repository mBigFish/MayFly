package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/service"
)

// ListCmdHistory 查询命令历史
// GET /api/v1/targets/:id/history?page=1&per_page=20
func ListCmdHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	page, perPage := ParsePage(c)
	history, total, err := service.ListCmdHistory(database.Get(), uint(id), page, perPage)
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OKPage(c, history, total, page, perPage)
}

// ClearCmdHistory 清空命令历史
// DELETE /api/v1/targets/:id/history
func ClearCmdHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteCmdHistory(database.Get(), uint(id)); err != nil {
		Fail(c, 500, "清空失败: "+err.Error())
		return
	}
	OKMsg(c, "已清空", nil)
}

// 确保 model 包被引用
var _ = model.CmdHistory{}
