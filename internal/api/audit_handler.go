package api

import (
	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/service"
)

// ListAuditLogs 查询审计日志
// GET /api/v1/audit?keyword=xxx&page=1&per_page=20
func ListAuditLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	page, perPage := ParsePage(c)
	logs, total, err := service.ListAuditLogs(database.Get(), keyword, page, perPage)
	if err != nil {
		Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	OKPage(c, logs, total, page, perPage)
}
