package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/user/rpc/internal/config"
	"github.com/aether-defense-system/service/user/rpc/internal/svc"
)

func TestGetUserLogic_GetUser(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	resp, err := logic.GetUser(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
}
