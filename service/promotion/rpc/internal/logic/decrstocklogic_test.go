package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/promotion/rpc/internal/config"
	"github.com/aether-defense-system/service/promotion/rpc/internal/svc"
)

func TestDecrStockLogic_DecrStock(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	resp, err := logic.DecrStock(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
}
