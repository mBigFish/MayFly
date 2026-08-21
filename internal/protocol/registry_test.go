package protocol

import (
	"context"
	"testing"

	"github.com/webshell-manager/webshell-manager/internal/operation"
	"github.com/webshell-manager/webshell-manager/internal/target"
)

// fakeProtocol 用于测试注册表。
type fakeProtocol struct {
	name string
}

func (f *fakeProtocol) Name() string { return f.name }
func (f *fakeProtocol) Check(_ context.Context, _ *target.Target) error {
	return nil
}
func (f *fakeProtocol) Execute(_ context.Context, _ *target.Target, _ *operation.Operation) (*operation.Result, error) {
	return &operation.Result{Success: true}, nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeProtocol{name: "php"})
	r.Register(&fakeProtocol{name: "jsp"})

	p, err := r.Get("php")
	if err != nil {
		t.Fatalf("Get(php) 失败: %v", err)
	}
	if p.Name() != "php" {
		t.Errorf("期望 php，得到 %q", p.Name())
	}
}

func TestRegistryGetUnknown(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Get("unknown"); err == nil {
		t.Error("获取未注册协议应报错")
	}
}

func TestRegistryNames(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeProtocol{name: "php"})
	r.Register(&fakeProtocol{name: "aspx"})

	names := r.Names()
	if len(names) != 2 {
		t.Errorf("期望 2 个协议，得到 %d", len(names))
	}
}
