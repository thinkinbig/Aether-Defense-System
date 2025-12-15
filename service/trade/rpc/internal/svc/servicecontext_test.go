package svc

import (
	"testing"

	"github.com/aether-defense-system/service/trade/rpc/internal/config"
)

func TestNewServiceContext(t *testing.T) {
	cfg := &config.Config{}
	ctx := NewServiceContext(cfg)

	if ctx == nil {
		t.Fatalf("expected non-nil ServiceContext")
	}
	if ctx.Config != cfg {
		t.Fatalf("expected Config pointer to be preserved")
	}
}
