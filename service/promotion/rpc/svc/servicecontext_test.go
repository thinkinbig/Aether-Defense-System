package svc

import (
	"testing"

	"github.com/aether-defense-system/service/promotion/rpc/config-internal"
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
	// When inventoryRedis is not configured, Redis client should remain nil
	// (unit tests should not require external Redis).
	if ctx.Redis != nil {
		t.Fatalf("expected Redis to be nil when inventoryRedis is not configured")
	}
}
