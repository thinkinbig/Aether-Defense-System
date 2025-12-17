//go:build integration
// +build integration

// Package promotion contains integration tests for promotion service.
// These tests verify the complete flow through gRPC service interface.
package promotion

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	commonredis "github.com/aether-defense-system/common/redis"
	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/test/integration"
)

// TestIntegration_StockDeduction_RedisLuaScript tests the Redis Lua script directly.
// This verifies the data layer integration (Redis + Lua script).
func TestIntegration_StockDeduction_RedisLuaScript(t *testing.T) {
	// Arrange: Set up test environment
	env := integration.SetupTestEnvironment(t)
	ctx := env.Context()

	// Set initial inventory
	courseID := int64(2001)
	initialStock := int64(100)
	deductionAmount := int64(25)

	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)
	err := env.Redis.Set(ctx, inventoryKey, initialStock, 0)
	require.NoError(t, err, "Failed to set initial inventory")

	// Verify initial stock
	stockBefore, err := env.Redis.Get(ctx, inventoryKey)
	require.NoError(t, err)
	assert.Equal(t, "100", stockBefore, "Initial stock should be 100")

	// Act: Execute inventory deduction using Redis client
	err = env.Redis.DecrStock(ctx, inventoryKey, deductionAmount)

	// Assert: Verify no error
	require.NoError(t, err, "DecrStock should not return error")

	// Assert: Verify stock was actually decremented in Redis
	stockAfter, err := env.Redis.Get(ctx, inventoryKey)
	require.NoError(t, err, "Failed to get stock after deduction")
	expectedStock := initialStock - deductionAmount
	assert.Equal(t, fmt.Sprintf("%d", expectedStock), stockAfter,
		"Stock should be decremented from %d to %d", initialStock, expectedStock)
}

// TestIntegration_StockDeduction_InsufficientStock_RedisLuaScript tests insufficient stock scenario.
func TestIntegration_StockDeduction_InsufficientStock_RedisLuaScript(t *testing.T) {
	// Arrange: Set up test environment
	env := integration.SetupTestEnvironment(t)
	ctx := env.Context()

	// Set initial inventory (less than deduction amount)
	courseID := int64(2002)
	initialStock := int64(10)
	deductionAmount := int64(50) // More than available

	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)
	err := env.Redis.Set(ctx, inventoryKey, initialStock, 0)
	require.NoError(t, err)

	// Act: Execute inventory deduction
	err = env.Redis.DecrStock(ctx, inventoryKey, deductionAmount)

	// Assert: Verify error
	require.Error(t, err, "DecrStock should return error for insufficient stock")
	assert.Contains(t, err.Error(), "Insufficient inventory", "Error message should indicate insufficient inventory")

	// Assert: Verify stock was NOT decremented
	stockAfter, err := env.Redis.Get(ctx, inventoryKey)
	require.NoError(t, err)
	assert.Equal(t, "10", stockAfter, "Stock should remain unchanged")
}

// TestIntegration_StockDeduction_NonExistentKey_RedisLuaScript tests non-existent key scenario.
func TestIntegration_StockDeduction_NonExistentKey_RedisLuaScript(t *testing.T) {
	// Arrange: Set up test environment
	env := integration.SetupTestEnvironment(t)
	ctx := env.Context()

	// Don't set inventory key (non-existent)
	courseID := int64(2003)
	deductionAmount := int64(10)

	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)

	// Act: Execute inventory deduction
	err := env.Redis.DecrStock(ctx, inventoryKey, deductionAmount)

	// Assert: Verify error
	require.Error(t, err, "DecrStock should return error for non-existent key")
	assert.Contains(t, err.Error(), "Inventory Key does not exist", "Error message should indicate key does not exist")
}

// TestIntegration_StockDeduction_MultipleDeductions_RedisLuaScript tests multiple sequential deductions.
func TestIntegration_StockDeduction_MultipleDeductions_RedisLuaScript(t *testing.T) {
	// Arrange: Set up test environment
	env := integration.SetupTestEnvironment(t)
	ctx := env.Context()

	// Set initial inventory
	courseID := int64(2004)
	initialStock := int64(100)

	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)
	err := env.Redis.Set(ctx, inventoryKey, initialStock, 0)
	require.NoError(t, err)

	// Act: Execute multiple deductions
	deductions := []int64{10, 20, 15}
	expectedStock := initialStock

	for _, amount := range deductions {
		deductErr := env.Redis.DecrStock(ctx, inventoryKey, amount)
		require.NoError(t, deductErr, "Deduction should succeed")
		expectedStock -= amount
	}

	// Assert: Verify final stock
	stockAfter, err := env.Redis.Get(ctx, inventoryKey)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%d", expectedStock), stockAfter,
		"Stock should be decremented correctly after multiple deductions")
}

// TestIntegration_StockDeduction_ConcurrentDeductions_RedisLuaScript tests concurrent inventory deductions.
// This verifies that the Lua script provides atomic operations.
func TestIntegration_StockDeduction_ConcurrentDeductions_RedisLuaScript(t *testing.T) {
	// Arrange: Set up test environment
	env := integration.SetupTestEnvironment(t)
	ctx := env.Context()

	// Set initial inventory
	courseID := int64(2005)
	initialStock := int64(100)
	deductionAmount := int64(10)
	concurrentRequests := 5

	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)
	err := env.Redis.Set(ctx, inventoryKey, initialStock, 0)
	require.NoError(t, err)

	// Act: Execute concurrent deductions
	done := make(chan bool, concurrentRequests)
	errorChan := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			deductErr := env.Redis.DecrStock(ctx, inventoryKey, deductionAmount)
			if deductErr != nil {
				errorChan <- deductErr
				done <- false
				return
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	successCount := 0
	for i := 0; i < concurrentRequests; i++ {
		select {
		case success := <-done:
			if success {
				successCount++
			}
		case deductErr := <-errorChan:
			t.Logf("Concurrent deduction error: %v", deductErr)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent deductions")
		}
	}

	// Assert: Verify all deductions succeeded
	assert.Equal(t, concurrentRequests, successCount, "All concurrent deductions should succeed")

	// Assert: Verify final stock
	expectedStock := initialStock - (int64(concurrentRequests) * deductionAmount)
	stockAfter, err := env.Redis.Get(ctx, inventoryKey)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%d", expectedStock), stockAfter,
		"Stock should be correctly decremented after concurrent operations")
}

// TestIntegration_StockDeduction_ThroughGRPC tests through gRPC service (requires running service).
// This test is skipped by default and should be run when the service is available.
func TestIntegration_StockDeduction_ThroughGRPC(t *testing.T) {
	// Skip if service is not running
	grpcAddr := os.Getenv("PROMOTION_RPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:8082"
	}
	if !isServiceRunning(grpcAddr) {
		t.Skipf("Promotion RPC service not running at %s. Start the service to run this test.", grpcAddr)
	}

	// This test writes to the SAME Redis DB used by the running service.
	// It intentionally does NOT use integration.SetupTestEnvironment because that helper FlushDB()s
	// the configured database, which would be dangerous if pointed at a live service DB.
	redisAddr := os.Getenv("PROMOTION_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB := 0
	if dbStr := os.Getenv("PROMOTION_REDIS_DB"); dbStr != "" {
		if _, err := fmt.Sscanf(dbStr, "%d", &redisDB); err != nil {
			t.Fatalf("invalid PROMOTION_REDIS_DB %q: %v", dbStr, err)
		}
	}

	svcRedis, err := commonredis.NewClient(&commonredis.Config{
		Addr:         redisAddr,
		Password:     "",
		DB:           redisDB,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	require.NoError(t, err, "Failed to connect to service Redis at %s (db=%d)", redisAddr, redisDB)
	t.Cleanup(func() {
		if closeErr := svcRedis.Close(); closeErr != nil {
			t.Logf("Warning: failed to close service Redis client: %v", closeErr)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to gRPC service
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to gRPC service")
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Logf("Warning: failed to close gRPC connection: %v", closeErr)
		}
	}()

	client := rpc.NewPromotionServiceClient(conn)

	// Set initial inventory
	courseID := int64(3001)
	initialStock := int64(100)
	deductionAmount := int64(25)

	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)
	err = svcRedis.Set(ctx, inventoryKey, initialStock, 0)
	require.NoError(t, err, "Failed to set initial inventory in service Redis")
	t.Cleanup(func() {
		// ctx might be canceled by the time Cleanup runs; use background context for best-effort cleanup.
		if delErr := svcRedis.Del(context.Background(), inventoryKey); delErr != nil {
			t.Logf("Warning: failed to delete test inventory key: %v", delErr)
		}
	})

	// Act: Call gRPC service
	req := &rpc.DecrStockRequest{
		CourseId: courseID,
		Num:      int32(deductionAmount),
	}

	resp, err := client.DecrStock(ctx, req)

	// Assert: Verify response
	require.NoError(t, err, "gRPC call should not return error")
	require.NotNil(t, resp, "Response should not be nil")
	assert.True(t, resp.Success, "Deduction should be successful")

	// Assert: Verify stock was actually decremented in Redis
	stockAfter, err := svcRedis.Get(ctx, inventoryKey)
	require.NoError(t, err)
	expectedStock := initialStock - deductionAmount
	assert.Equal(t, fmt.Sprintf("%d", expectedStock), stockAfter,
		"Stock should be decremented from %d to %d", initialStock, expectedStock)
}

// isServiceRunning checks if a gRPC service is running at the given address.
func isServiceRunning(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	if closeErr := conn.Close(); closeErr != nil {
		_ = closeErr // best-effort, ignore
	}
	return true
}
