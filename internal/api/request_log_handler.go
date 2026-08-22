package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/protocol"
	"mayfly/internal/service"
)

// ListRequestLogs 列出请求日志
// GET /api/v1/request-logs?target_id=1&keyword=xxx&page=1&page_size=20
func ListRequestLogs(c *gin.Context) {
	targetID, _ := strconv.ParseUint(c.Query("target_id"), 10, 64)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	logs, total, err := service.ListRequestLogs(database.Get(), uint(targetID), keyword, page, pageSize)
	if err != nil {
		Fail(c, 500, "查询失败")
		return
	}
	OKPage(c, logs, total, page, pageSize)
}

// GetRequestLog 获取请求日志详情
// GET /api/v1/request-logs/:id
func GetRequestLog(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	log, err := service.GetRequestLog(database.Get(), uint(id))
	if err != nil {
		Fail(c, 404, "日志不存在")
		return
	}
	OK(c, log)
}

// executeAndLog 执行操作并记录请求日志
func executeAndLog(ctx context.Context, c *gin.Context, target *model.Target, op *protocol.Operation) (*protocol.Result, error) {
	start := time.Now()

	result, err := protocol.ExecuteForTarget(ctx, target, op)
	duration := time.Since(start)

	// 构建请求参数 JSON
	paramsJSON, _ := json.Marshal(op.Params)

	// 构建请求日志
	log := &model.RequestLog{
		TargetID:   target.ID,
		TargetName: target.Name,
		UserID:     c.GetUint("user_id"),
		Username:   c.GetString("username"),
		Operation:  string(op.Type),
		Params:     string(paramsJSON),
		Duration:   duration.Milliseconds(),
	}

	if err != nil {
		log.Status = "error"
		log.Error = err.Error()
	} else if result != nil {
		log.Status = result.Status
		log.Response = string(result.Data)
		if result.Message != "" {
			log.Error = result.Message
		}
	} else {
		log.Status = "error"
		log.Error = "nil result"
	}

	// 异步记录日志
	go service.LogRequest(database.Get(), log)

	return result, err
}
