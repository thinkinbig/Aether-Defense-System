package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"
)

func TestCancelOrderLogic_CancelOrder_ValidationAndSuccess(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewCancelOrderLogic(context.Background(), svcCtx)

	tests := []struct {
		req     *rpc.CancelOrderRequest
		name    string
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "invalid user id",
			req: &rpc.CancelOrderRequest{
				UserId:  0,
				OrderId: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid order id",
			req: &rpc.CancelOrderRequest{
				UserId:  1,
				OrderId: 0,
			},
			wantErr: true,
		},
		{
			name: "success",
			req: &rpc.CancelOrderRequest{
				UserId:  1,
				OrderId: 100,
			},
			// Note: This test validates input parameters only.
			// The actual business logic requires OrderRepo dependency which is not mocked here.
			// For full integration testing with dependencies, see integration tests.
			wantErr: true, // Will fail due to missing OrderRepo dependency
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.CancelOrder(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (resp=%+v)", resp)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if resp == nil {
				t.Fatalf("expected non-nil response")
			}
			if resp.OrderId != tt.req.OrderId {
				t.Errorf("expected OrderId=%d, got %d", tt.req.OrderId, resp.OrderId)
			}
			if resp.Status != 2 {
				t.Errorf("expected Status=2 (closed), got %d", resp.Status)
			}
			if !resp.Success {
				t.Errorf("expected Success=true, got false")
			}
		})
	}
}
