package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"
)

func TestPlaceOrderLogic_PlaceOrder_ValidationAndSuccess(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	tests := []struct {
		req     *rpc.PlaceOrderRequest
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
			req: &rpc.PlaceOrderRequest{
				UserId:     0,
				OrderId:    1,
				CourseIds:  []int64{1},
				RealAmount: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid order id",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    0,
				CourseIds:  []int64{1},
				RealAmount: 100,
			},
			wantErr: true,
		},
		{
			name: "empty course ids",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    1,
				CourseIds:  []int64{},
				RealAmount: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid real amount",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    1,
				CourseIds:  []int64{1},
				RealAmount: 0,
			},
			wantErr: true,
		},
		{
			name: "success",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    100,
				CourseIds:  []int64{1, 2},
				CouponIds:  []int64{10},
				RealAmount: 1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.PlaceOrder(tt.req)

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
			if resp.PayAmount != tt.req.RealAmount {
				t.Errorf("expected PayAmount=%d, got %d", tt.req.RealAmount, resp.PayAmount)
			}
			if resp.Status != 1 {
				t.Errorf("expected Status=1 (pending payment), got %d", resp.Status)
			}
		})
	}
}
