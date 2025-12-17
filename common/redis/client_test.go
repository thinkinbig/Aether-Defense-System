package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

// TestRedisClient tests require a running Redis instance.
// These tests can be run with: go test -tags=integration

const (
	testValue1 = "value1"
	testLocked = "locked"
)

func setupTestClient(t *testing.T) *Client {
	config := &Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           15, // Use DB 15 for testing to avoid conflicts
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}

	// Clean up test database
	ctx := context.Background()
	if err := client.rdb.FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: failed to flush test database: %v", err)
	}

	return client
}

func TestNewClient(t *testing.T) {
	config := DefaultConfig()
	config.DB = 15 // Use test database

	client, err := NewClient(config)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	// Test ping
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestClient_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()

	// Test Set and Get
	err := client.Set(ctx, "test:key", "test:value", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	value, err := client.Get(ctx, "test:key")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if value != "test:value" {
		t.Errorf("Get() = %v, want %v", value, "test:value")
	}

	// Test Exists
	count, err := client.Exists(ctx, "test:key")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Exists() = %v, want 1", count)
	}

	// Test Del
	err = client.Del(ctx, "test:key")
	if err != nil {
		t.Errorf("Del() error = %v", err)
	}

	// Verify deletion
	_, getErr := client.Get(ctx, "test:key")
	if !errors.Is(getErr, redisv9.Nil) {
		t.Errorf("Expected redisv9.Nil after deletion, got %v", getErr)
	}
}

func TestClient_HashOperations(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()

	// Test HSet
	err := client.HSet(ctx, "test:hash", "field1", "value1", "field2", "value2")
	if err != nil {
		t.Errorf("HSet() error = %v", err)
	}

	// Test HGet
	value, err := client.HGet(ctx, "test:hash", "field1")
	if err != nil {
		t.Errorf("HGet() error = %v", err)
	}
	if value != testValue1 {
		t.Errorf("HGet() = %v, want %v", value, testValue1)
	}

	// Test HGetAll
	allFields, err := client.HGetAll(ctx, "test:hash")
	if err != nil {
		t.Errorf("HGetAll() error = %v", err)
	}
	if len(allFields) != 2 {
		t.Errorf("HGetAll() returned %d fields, want 2", len(allFields))
	}
	if allFields["field1"] != testValue1 || allFields["field2"] != "value2" {
		t.Errorf("HGetAll() = %v, want field1:%s, field2:value2", allFields, testValue1)
	}
}

func TestClient_SetOperations(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()

	// Test SAdd
	err := client.SAdd(ctx, "test:set", "member1", "member2")
	if err != nil {
		t.Errorf("SAdd() error = %v", err)
	}

	// Test SIsMember
	exists, err := client.SIsMember(ctx, "test:set", "member1")
	if err != nil {
		t.Errorf("SIsMember() error = %v", err)
	}
	if !exists {
		t.Errorf("SIsMember() = %v, want true", exists)
	}

	// Test non-existent member
	exists, err = client.SIsMember(ctx, "test:set", "nonexistent")
	if err != nil {
		t.Errorf("SIsMember() error = %v", err)
	}
	if exists {
		t.Errorf("SIsMember() = %v, want false", exists)
	}
}

func TestClient_DecrStock(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()
	inventoryKey := "test:inventory:123"

	// Set initial inventory
	err := client.Set(ctx, inventoryKey, "100", time.Minute)
	if err != nil {
		t.Fatalf("Failed to set initial inventory: %v", err)
	}

	// Test successful deduction
	err = client.DecrStock(ctx, inventoryKey, 10)
	if err != nil {
		t.Errorf("DecrStock() error = %v", err)
	}

	// Verify remaining inventory
	remaining, err := client.Get(ctx, inventoryKey)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if remaining != "90" {
		t.Errorf("Remaining inventory = %v, want 90", remaining)
	}

	// Test insufficient inventory
	err = client.DecrStock(ctx, inventoryKey, 100)
	if err == nil {
		t.Errorf("DecrStock() expected error for insufficient inventory")
	}

	// Test non-existent key
	err = client.DecrStock(ctx, "nonexistent:key", 1)
	if err == nil {
		t.Errorf("DecrStock() expected error for non-existent key")
	}
}

func TestClient_DecrStockWithUser(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()
	inventoryKey := "test:inventory:456"
	userSetKey := "test:purchased:456"

	// Set initial inventory
	err := client.Set(ctx, inventoryKey, "50", time.Minute)
	if err != nil {
		t.Fatalf("Failed to set initial inventory: %v", err)
	}

	// Test first purchase by user
	err = client.DecrStockWithUser(ctx, inventoryKey, userSetKey, 1, 123)
	if err != nil {
		t.Errorf("DecrStockWithUser() error = %v", err)
	}

	// Verify inventory deduction
	remaining, err := client.Get(ctx, inventoryKey)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if remaining != "49" {
		t.Errorf("Remaining inventory = %v, want 49", remaining)
	}

	// Verify user is tracked
	purchased, err := client.SIsMember(ctx, userSetKey, 123)
	if err != nil {
		t.Errorf("SIsMember() error = %v", err)
	}
	if !purchased {
		t.Errorf("User should be marked as purchased")
	}

	// Test duplicate purchase by same user
	err = client.DecrStockWithUser(ctx, inventoryKey, userSetKey, 1, 123)
	if err == nil {
		t.Errorf("DecrStockWithUser() expected error for duplicate purchase")
	}

	// Verify inventory unchanged after duplicate attempt
	remaining, err = client.Get(ctx, inventoryKey)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if remaining != "49" {
		t.Errorf("Inventory should remain 49 after duplicate attempt, got %v", remaining)
	}
}

func TestClient_SetNXWithExpire(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()
	key := "test:lock:789"

	// Test setting new key
	err := client.SetNXWithExpire(ctx, key, testLocked, 10*time.Second)
	if err != nil {
		t.Errorf("SetNXWithExpire() error = %v", err)
	}

	// Verify key exists
	value, err := client.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if value != testLocked {
		t.Errorf("Get() = %v, want %s", value, testLocked)
	}

	// Test setting existing key (should fail)
	err = client.SetNXWithExpire(ctx, key, "locked2", 10*time.Second)
	if err == nil {
		t.Errorf("SetNXWithExpire() expected error for existing key")
	}

	// Verify original value unchanged
	value, err = client.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if value != testLocked {
		t.Errorf("Value should remain %s, got %v", testLocked, value)
	}
}

func TestClient_IncrWithExpire(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()
	key := "test:counter:999"

	// Test first increment (should set expiration)
	count, err := client.IncrWithExpire(ctx, key, 60*time.Second)
	if err != nil {
		t.Errorf("IncrWithExpire() error = %v", err)
	}
	if count != 1 {
		t.Errorf("IncrWithExpire() = %v, want 1", count)
	}

	// Test second increment (should not reset expiration)
	count, err = client.IncrWithExpire(ctx, key, 60*time.Second)
	if err != nil {
		t.Errorf("IncrWithExpire() error = %v", err)
	}
	if count != 2 {
		t.Errorf("IncrWithExpire() = %v, want 2", count)
	}

	// Verify key has expiration
	ttl := client.rdb.TTL(ctx, key).Val()
	if ttl <= 0 {
		t.Errorf("Key should have expiration, got TTL: %v", ttl)
	}
}

func TestClient_ConcurrentDecrStock(t *testing.T) {
	client := setupTestClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Warning: failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()
	inventoryKey := "test:concurrent:inventory"
	initialStock := int64(100)

	// Set initial inventory
	err := client.Set(ctx, inventoryKey, initialStock, time.Minute)
	if err != nil {
		t.Fatalf("Failed to set initial inventory: %v", err)
	}

	// Run concurrent deductions
	numGoroutines := 50
	deductionAmount := int64(1)

	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := client.DecrStock(ctx, inventoryKey, deductionAmount)
			errChan <- err
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		if err == nil {
			successCount++
		}
	}

	// Verify final inventory
	expectedFinalStock := initialStock - int64(successCount)

	// Use a more robust check by getting the actual integer value
	remaining, remainingErr := client.rdb.Get(ctx, inventoryKey).Int64()
	if remainingErr != nil {
		t.Errorf("Failed to get final inventory as int64: %v", remainingErr)
	} else if remaining != expectedFinalStock {
		t.Errorf("Final inventory = %d, want %d (successful deductions: %d)",
			remaining, expectedFinalStock, successCount)
	}

	// Ensure we had some successful operations
	if successCount == 0 {
		t.Errorf("Expected some successful deductions, got 0")
	}

	// Ensure we didn't over-deduct
	if successCount > int(initialStock) {
		t.Errorf("Too many successful deductions: %d, max should be %d", successCount, initialStock)
	}
}

func TestKeyNamingHelper(t *testing.T) {
	helper := NewKeyNamingHelper()

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{
			name:     "InventoryKey",
			method:   func() string { return helper.InventoryKey(123) },
			expected: "promotion:stock:123",
		},
		{
			name:     "UserPurchaseKey",
			method:   func() string { return helper.UserPurchaseKey(456) },
			expected: "promotion:purchased:456",
		},
		{
			name:     "OrderLockKey",
			method:   func() string { return helper.OrderLockKey(789) },
			expected: "trade:lock:789",
		},
		{
			name:     "UserSessionKey",
			method:   func() string { return helper.UserSessionKey(999) },
			expected: "user:session:999",
		},
		{
			name:     "RateLimitKey",
			method:   func() string { return helper.RateLimitKey(111, "login") },
			expected: "ratelimit:login:111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			if result != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Addr != "localhost:6379" {
		t.Errorf("DefaultConfig().Addr = %v, want localhost:6379", config.Addr)
	}
	if config.PoolSize != 100 {
		t.Errorf("DefaultConfig().PoolSize = %v, want 100", config.PoolSize)
	}
	if config.MinIdleConns != 10 {
		t.Errorf("DefaultConfig().MinIdleConns = %v, want 10", config.MinIdleConns)
	}
}
