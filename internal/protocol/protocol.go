// Package protocol 定义协议抽象与注册表。
// 协议层与 HTTP Transport 解耦：协议不直接依赖 http.Client，必须通过 Transport 通信。
package protocol

import (
	"context"

	"github.com/webshell-manager/webshell-manager/internal/operation"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// Protocol 是统一协议接口。
type Protocol interface {
	// Name 返回协议名称（如 php、jsp、aspx）。
	Name() string

	// Check 校验目标是否可用（探活）。
	Check(ctx context.Context, t *target.Target) error

	// Execute 在目标上执行统一操作。
	Execute(ctx context.Context, t *target.Target, op *operation.Operation) (*operation.Result, error)
}
