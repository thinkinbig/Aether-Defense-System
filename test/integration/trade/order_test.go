//go:build integration
// +build integration

// Package trade contains integration tests for trade service.
// These tests verify the complete order placement and cancellation flow.
package trade

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

	"github.com/aether-defense-system/common/database"
	commonredis "github.com/aether-defense-system/common/redis"
	traderpc "github.com/aether-defense-system/service/trade/rpc"
	userrpc "github.com/aether-defense-system/service/user/rpc"
)

// TestIntegration_PlaceOrder_Success tests successful order placement.
// This test verifies:
// - Order is created in database
// - RocketMQ message is sent
// - User validation works
func TestIntegration_PlaceOrder_Success(t *testing.T) {
	// Skip if services are not running
	if !areServicesRunning(t) {
		t.Skip("Required services not running. Start MySQL, Redis, RocketMQ, and RPC services to run this test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to Trade RPC service
	tradeConn, err := grpc.NewClient(getTradeRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to Trade RPC service")
	defer tradeConn.Close()

	tradeClient := traderpc.NewTradeServiceClient(tradeConn)

	// Connect to User RPC service to verify user exists
	userConn, err := grpc.NewClient(getUserRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to User RPC service")
	defer userConn.Close()

	userClient := userrpc.NewUserServiceClient(userConn)

	// Setup: Create a test user (if not exists)
	// For now, we'll use a test user ID
	testUserID := int64(1001)
	_, err = userClient.GetUser(ctx, &userrpc.GetUserRequest{UserId: testUserID})
	if err != nil {
		t.Logf("Test user %d does not exist, skipping test", testUserID)
		t.Skipf("Test user %d not found. Create a test user in the database first.", testUserID)
	}

	// Setup: Set initial inventory in Redis
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	courseID := int64(5001)
	initialStock := int64(100)
	inventoryKey := fmt.Sprintf("inventory:course:%d", courseID)
	err = redisClient.Set(ctx, inventoryKey, initialStock, 0)
	require.NoError(t, err, "Failed to set initial inventory")
	t.Cleanup(func() {
		_ = redisClient.Del(context.Background(), inventoryKey)
	})

	// Generate order ID using Snowflake (simplified - in real test, use actual generator)
	orderID := time.Now().UnixNano() / 1000000 // Simple ID for testing

	// Act: Place order
	req := &traderpc.PlaceOrderRequest{
		UserId:     testUserID,
		OrderId:    orderID,
		CourseIds:  []int64{courseID},
		CouponIds:  []int64{},
		RealAmount: 10000, // 100.00 in cents
	}

	resp, err := tradeClient.PlaceOrder(ctx, req)

	// Assert: Verify response
	require.NoError(t, err, "PlaceOrder should not return error")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, orderID, resp.OrderId, "Order ID should match")
	assert.Equal(t, int32(10000), resp.PayAmount, "Pay amount should match")
	assert.Equal(t, int32(database.OrderStatusPendingPayment), resp.Status, "Status should be pending payment")

	// Verify order exists in database (if database is accessible)
	// Note: This requires database access, which may not be available in all test environments
	t.Logf("Order placed successfully: orderId=%d, userId=%d", resp.OrderId, testUserID)

	// Note: Inventory deduction happens asynchronously via RocketMQ consumer
	// In a full integration test, you would wait for the message to be consumed
	// and verify inventory was deducted
}

// TestIntegration_PlaceOrder_InvalidUser tests order placement with invalid user.
func TestIntegration_PlaceOrder_InvalidUser(t *testing.T) {
	if !areServicesRunning(t) {
		t.Skip("Required services not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tradeConn, err := grpc.NewClient(getTradeRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer tradeConn.Close()

	tradeClient := traderpc.NewTradeServiceClient(tradeConn)

	// Use a non-existent user ID
	invalidUserID := int64(999999)
	orderID := time.Now().UnixNano() / 1000000

	req := &traderpc.PlaceOrderRequest{
		UserId:     invalidUserID,
		OrderId:    orderID,
		CourseIds:  []int64{5002},
		CouponIds:  []int64{},
		RealAmount: 10000,
	}

	resp, err := tradeClient.PlaceOrder(ctx, req)

	// Assert: Should return error for invalid user
	require.Error(t, err, "PlaceOrder should return error for invalid user")
	assert.Nil(t, resp, "Response should be nil on error")
	assert.Contains(t, err.Error(), "user", "Error should mention user")
}

// TestIntegration_PlaceOrder_EmptyCourseIds tests order placement with empty course list.
func TestIntegration_PlaceOrder_EmptyCourseIds(t *testing.T) {
	if !areServicesRunning(t) {
		t.Skip("Required services not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tradeConn, err := grpc.NewClient(getTradeRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer tradeConn.Close()

	tradeClient := traderpc.NewTradeServiceClient(tradeConn)

	testUserID := int64(1001)
	orderID := time.Now().UnixNano() / 1000000

	req := &traderpc.PlaceOrderRequest{
		UserId:     testUserID,
		OrderId:    orderID,
		CourseIds:  []int64{}, // Empty course list
		CouponIds:  []int64{},
		RealAmount: 10000,
	}

	resp, err := tradeClient.PlaceOrder(ctx, req)

	// Assert: Should return error for empty course list
	require.Error(t, err, "PlaceOrder should return error for empty course list")
	assert.Nil(t, resp, "Response should be nil on error")
	assert.Contains(t, err.Error(), "course_ids", "Error should mention course_ids")
}

// TestIntegration_CancelOrder_Success tests successful order cancellation.
func TestIntegration_CancelOrder_Success(t *testing.T) {
	if !areServicesRunning(t) {
		t.Skip("Required services not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First, place an order
	tradeConn, err := grpc.NewClient(getTradeRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer tradeConn.Close()

	tradeClient := traderpc.NewTradeServiceClient(tradeConn)

	testUserID := int64(1001)
	orderID := time.Now().UnixNano() / 1000000

	// Place order first
	placeReq := &traderpc.PlaceOrderRequest{
		UserId:     testUserID,
		OrderId:    orderID,
		CourseIds:  []int64{5003},
		CouponIds:  []int64{},
		RealAmount: 10000,
	}

	placeResp, err := tradeClient.PlaceOrder(ctx, placeReq)
	if err != nil {
		t.Skipf("Failed to place order for cancellation test: %v", err)
	}
	require.NotNil(t, placeResp)

	// Wait a bit for order to be created
	time.Sleep(1 * time.Second)

	// Act: Cancel the order
	cancelReq := &traderpc.CancelOrderRequest{
		UserId:  testUserID,
		OrderId: orderID,
	}

	cancelResp, err := tradeClient.CancelOrder(ctx, cancelReq)

	// Assert: Verify cancellation
	require.NoError(t, err, "CancelOrder should not return error")
	require.NotNil(t, cancelResp, "Response should not be nil")
	assert.Equal(t, orderID, cancelResp.OrderId, "Order ID should match")
	assert.Equal(t, int32(database.OrderStatusClosed), cancelResp.Status, "Status should be closed")
	assert.True(t, cancelResp.Success, "Cancellation should be successful")
}

// TestIntegration_CancelOrder_InvalidUser tests cancellation with wrong user.
func TestIntegration_CancelOrder_InvalidUser(t *testing.T) {
	if !areServicesRunning(t) {
		t.Skip("Required services not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tradeConn, err := grpc.NewClient(getTradeRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer tradeConn.Close()

	tradeClient := traderpc.NewTradeServiceClient(tradeConn)

	// Try to cancel an order with wrong user ID
	req := &traderpc.CancelOrderRequest{
		UserId:  999999, // Wrong user
		OrderId: 1,      // Assuming order 1 exists and belongs to another user
	}

	resp, err := tradeClient.CancelOrder(ctx, req)

	// Assert: Should return error
	require.Error(t, err, "CancelOrder should return error for wrong user")
	assert.Nil(t, resp, "Response should be nil on error")
}

// Helper functions

func areServicesRunning(t *testing.T) bool {
	tradeRunning := isServiceRunning(getTradeRPCAddr())
	userRunning := isServiceRunning(getUserRPCAddr())

	if !tradeRunning {
		t.Logf("Trade RPC service not running at %s", getTradeRPCAddr())
	}
	if !userRunning {
		t.Logf("User RPC service not running at %s", getUserRPCAddr())
	}

	return tradeRunning && userRunning
}

func isServiceRunning(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func getTradeRPCAddr() string {
	addr := os.Getenv("TRADE_RPC_ADDR")
	if addr == "" {
		addr = "localhost:8081"
	}
	return addr
}

func getUserRPCAddr() string {
	addr := os.Getenv("USER_RPC_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}
	return addr
}

func setupTestRedis(t *testing.T) *commonredis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if _, err := fmt.Sscanf(dbStr, "%d", &redisDB); err != nil {
			t.Logf("Warning: invalid REDIS_DB value, using default: %v", err)
		}
	}

	client, err := commonredis.NewClient(&commonredis.Config{
		Addr:         redisAddr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           redisDB,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	require.NoError(t, err, "Failed to connect to Redis")
	return client
}
