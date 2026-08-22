package adapters

import (
	"context"
	"fmt"

	"mayfly/internal/model"
	"mayfly/internal/protocol"
	"mayfly/internal/transport"
)

// ASPProtocol ASP WebShell 协议适配器
type ASPProtocol struct {
	transport transport.Transport
}

func init() {
	protocol.GetRegistry().Register(&ASPProtocol{
		transport: transport.NewHTTPTransport(),
	})
}

func (p *ASPProtocol) Name() string { return "asp" }

func (p *ASPProtocol) Check(ctx context.Context, target *model.Target) error {
	result, err := p.Execute(ctx, target, &protocol.Operation{
		Type:   protocol.OpSysInfo,
		Params: map[string]string{},
	})
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("连接失败: %s", result.Message)
	}
	return nil
}

func (p *ASPProtocol) Execute(ctx context.Context, target *model.Target, op *protocol.Operation) (*protocol.Result, error) {
	return aspProtocolExecute(ctx, p.transport, target, op)
}
