package adapters

import (
	"context"
	"fmt"

	"mayfly/internal/model"
	"mayfly/internal/protocol"
	"mayfly/internal/transport"
)

// JSPProtocol JSP WebShell 协议适配器
type JSPProtocol struct {
	transport transport.Transport
}

func init() {
	protocol.GetRegistry().Register(&JSPProtocol{
		transport: transport.NewHTTPTransport(),
	})
}

func (p *JSPProtocol) Name() string { return "jsp" }

func (p *JSPProtocol) Check(ctx context.Context, target *model.Target) error {
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

func (p *JSPProtocol) Execute(ctx context.Context, target *model.Target, op *protocol.Operation) (*protocol.Result, error) {
	return base64ProtocolExecute(ctx, p.transport, target, op)
}
