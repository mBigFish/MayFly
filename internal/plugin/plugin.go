package plugin

import (
	"context"
	"mayfly/internal/model"
)

// Result 插件执行结果
type Result struct {
	Status  string      `json:"status"`  // ok / error
	Data    interface{} `json:"data"`    // 结构化数据
	Message string      `json:"message"` // 错误信息或描述
}

// Plugin 插件接口
type Plugin interface {
	Name() string
	Version() string
	Description() string
	Execute(ctx context.Context, target *model.Target, params map[string]string) (*Result, error)
}
