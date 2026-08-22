package adapters

import (
	"context"
	"fmt"

	"mayfly/internal/model"
	"mayfly/internal/protocol"
	"mayfly/internal/transport"
)

// ASPXProtocol ASPX WebShell 协议适配器
type ASPXProtocol struct {
	transport transport.Transport
}

func init() {
	protocol.GetRegistry().Register(&ASPXProtocol{
		transport: transport.NewHTTPTransport(),
	})
}

func (p *ASPXProtocol) Name() string { return "aspx" }

func (p *ASPXProtocol) Check(ctx context.Context, target *model.Target) error {
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

func (p *ASPXProtocol) Execute(ctx context.Context, target *model.Target, op *protocol.Operation) (*protocol.Result, error) {
	return base64ProtocolExecute(ctx, p.transport, target, op)
}
