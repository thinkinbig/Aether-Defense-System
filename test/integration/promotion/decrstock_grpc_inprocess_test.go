//go:build integration
// +build integration

package promotion

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/config-internal"
)

// This test is the "moved" version of the previously-added integration test that directly imported
// internal/logic. Per Go's internal package rules, tests under test/integration cannot import
// service/.../internal/* packages. Instead, we validate the same behavior through an in-process gRPC
// server, which exercises internal/logic via the generated server implementation.
func TestIntegration_StockDeduction_GrpcInProcess(t *testing.T) {
	env := integration.SetupTestEnvironment(t)
	redisCtx := env.Context()

	courseID := int64(4001)
	initialStock := int64(100)
	deductionAmount := int64(2)
	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)

	err := env.Redis.Set(redisCtx, inventoryKey, initialStock, 0)
	require.NoError(t, err, "failed to set initial inventory")
	t.Cleanup(func() {
		_ = env.Redis.Del(context.Background(), inventoryKey)
	})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "failed to listen")
	t.Cleanup(func() { _ = lis.Close() })

	grpcServer := grpc.NewServer()
	t.Cleanup(grpcServer.Stop)

	// Create ServiceContext directly since we can't import internal/config
	// The test only needs Redis, so we can set Config to nil
	svcCtx := &svc.ServiceContext{
		Config: nil,
		Redis:  env.Redis,
	}
	rpc.RegisterPromotionServiceServer(grpcServer, server.NewPromotionServiceServer(svcCtx))

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	callCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "failed to connect to in-process gRPC server")
	t.Cleanup(func() { _ = conn.Close() })

	client := rpc.NewPromotionServiceClient(conn)

	resp, err := client.DecrStock(callCtx, &rpc.DecrStockRequest{
		CourseId: courseID,
		Num:      int32(deductionAmount),
	})
	require.NoError(t, err, "gRPC call should not error")
	require.NotNil(t, resp)
	assert.True(t, resp.Success)

	stockAfter, err := env.Redis.Get(redisCtx, inventoryKey)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%d", initialStock-deductionAmount), stockAfter)
}
