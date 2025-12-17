package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/common/redis"
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
		{
			name: "success",
			req: &rpc.DecrStockRequest{
				CourseId: 1,
				Num:      2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For success test, try to initialize Redis client if available
			if !tt.wantErr && tt.name == "success" {
				redisClient, err := redis.NewClient(redis.DefaultConfig())
				if err != nil {
					t.Skipf("Redis not available for integration test: %v", err)
				}
				defer func() {
					if closeErr := redisClient.Close(); closeErr != nil {
						t.Logf("Warning: failed to close Redis client: %v", closeErr)
					}
				}()
				svcCtx.Redis = redisClient

				// Set initial inventory for the test
				ctx := context.Background()
				inventoryKey := "inventory:course:1"
				err = redisClient.Set(ctx, inventoryKey, "100", 0) // 0 = no expiration for test
				if err != nil {
					t.Fatalf("Failed to set initial inventory: %v", err)
				}
				// Clean up after test
				defer func() {
					if delErr := redisClient.Del(ctx, inventoryKey); delErr != nil {
						t.Logf("Warning: failed to delete test inventory key: %v", delErr)
					}
				}()
			}

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
