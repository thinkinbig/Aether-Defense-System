package logic

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/internal/config"
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

type fakeInventoryRedis struct {
	store        map[string]int64
	getBeforeErr error
	getAfterErr  error
	decrErr      error

	getCall int
}

func (f *fakeInventoryRedis) Get(_ context.Context, key string) (string, error) {
	f.getCall++

	if f.getCall == 1 && f.getBeforeErr != nil {
		return "", f.getBeforeErr
	}
	if f.getCall >= 2 && f.getAfterErr != nil {
		return "", f.getAfterErr
	}

	if f.store == nil {
		return "", fmt.Errorf("store not initialized")
	}
	v, ok := f.store[key]
	if !ok {
		return "", fmt.Errorf("key not found")
	}
	return strconv.FormatInt(v, 10), nil
}

func (f *fakeInventoryRedis) DecrStock(_ context.Context, inventoryKey string, quantity int64) error {
	if f.decrErr != nil {
		return f.decrErr
	}
	if f.store == nil {
		return fmt.Errorf("store not initialized")
	}
	cur, ok := f.store[inventoryKey]
	if !ok {
		return fmt.Errorf("key not found")
	}
	f.store[inventoryKey] = cur - quantity
	return nil
}

func TestDecrStockLogic_DecrStock_NoRedisConfigured(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg, Redis: nil}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	_, err := logic.DecrStock(&rpc.DecrStockRequest{CourseId: 1, Num: 1})
	if err == nil {
		t.Fatalf("expected error when Redis is nil")
	}
}

func TestDecrStockLogic_DecrStock_Success_Unit(t *testing.T) {
	cfg := &config.Config{}
	fake := &fakeInventoryRedis{
		store: map[string]int64{
			"inventory:course:1": 100,
		},
	}
	svcCtx := &svc.ServiceContext{Config: cfg, Redis: fake}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	resp, err := logic.DecrStock(&rpc.DecrStockRequest{CourseId: 1, Num: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}
	if got := fake.store["inventory:course:1"]; got != 98 {
		t.Fatalf("expected stock=98, got %d", got)
	}
}

func TestDecrStockLogic_DecrStock_DecrError(t *testing.T) {
	cfg := &config.Config{}
	fake := &fakeInventoryRedis{
		store: map[string]int64{
			"inventory:course:1": 100,
		},
		decrErr: fmt.Errorf("boom"),
	}
	svcCtx := &svc.ServiceContext{Config: cfg, Redis: fake}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	resp, err := logic.DecrStock(&rpc.DecrStockRequest{CourseId: 1, Num: 2})
	if err != nil {
		t.Fatalf("expected no error (business failure encoded in response), got %v", err)
	}
	if resp == nil || resp.Success {
		t.Fatalf("expected failure response, got %+v", resp)
	}
}

func TestDecrStockLogic_DecrStock_GetErrorsStillSucceed(t *testing.T) {
	cfg := &config.Config{}
	fake := &fakeInventoryRedis{
		store: map[string]int64{
			"inventory:course:1": 100,
		},
		getBeforeErr: fmt.Errorf("before get failed"),
		getAfterErr:  fmt.Errorf("after get failed"),
	}
	svcCtx := &svc.ServiceContext{Config: cfg, Redis: fake}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	resp, err := logic.DecrStock(&rpc.DecrStockRequest{CourseId: 1, Num: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}
}
