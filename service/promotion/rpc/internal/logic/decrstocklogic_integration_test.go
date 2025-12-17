//go:build integration

package logic

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aether-defense-system/common/redis"
	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/config-internal"
	"github.com/aether-defense-system/service/promotion/rpc/svc"
)

func TestDecrStockLogic_DecrStock_Success_Integration(t *testing.T) {
	addr := redisAddrFromEnv()

	rcfg := redis.DefaultConfig()
	rcfg.Addr = addr

	redisClient, err := redis.NewClient(rcfg)
	if err != nil {
		t.Fatalf("Redis not available at %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg, Redis: redisClient}
	logic := NewDecrStockLogic(context.Background(), svcCtx)

	ctx := context.Background()
	inventoryKey := "inventory:course:1"

	if err := redisClient.Set(ctx, inventoryKey, "100", 0); err != nil {
		t.Fatalf("failed to set initial inventory: %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Del(ctx, inventoryKey) })

	resp, err := logic.DecrStock(&rpc.DecrStockRequest{CourseId: 1, Num: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}

	after, err := redisClient.Get(ctx, inventoryKey)
	if err != nil {
		t.Fatalf("failed to read inventory after deduction: %v", err)
	}
	if after != "98" {
		t.Fatalf("expected stock=98 after deduction, got %s", after)
	}
}

func redisAddrFromEnv() string {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "6379"
	}
	return fmt.Sprintf("%s:%s", host, port)
}
