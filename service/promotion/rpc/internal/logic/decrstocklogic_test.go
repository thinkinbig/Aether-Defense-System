package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/config-internal"
	"github.com/aether-defense-system/service/promotion/rpc/svc"
)

func TestDecrStockLogic_DecrStock_ValidationAndSuccess(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	tests := []struct {
		req     *rpc.DecrStockRequest
		name    string
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "invalid course id",
			req: &rpc.DecrStockRequest{
				CourseId: 0,
				Num:      1,
			},
			wantErr: true,
		},
		{
			name: "invalid num",
			req: &rpc.DecrStockRequest{
				CourseId: 1,
				Num:      0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.DecrStock(tt.req)

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
			if !resp.Success {
				t.Errorf("expected Success=true, got false")
			}
			if resp.Message == "" {
				t.Errorf("expected non-empty Message")
			}
		})
	}
}
