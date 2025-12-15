package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"
)

func TestPlaceOrderLogic_PlaceOrder(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	resp, err := logic.PlaceOrder(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
}
