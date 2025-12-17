//go:build integration
// +build integration

// Package integration provides setup utilities for integration tests.
package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aether-defense-system/common/redis"
)

// TestEnvironment holds the test environment setup including clients and services.
type TestEnvironment struct {
	Redis *redis.Client
	ctx   context.Context
}

// SetupTestEnvironment creates a test environment with real dependencies.
// It uses environment variables to configure connections:
//   - REDIS_ADDR: Redis server address (default: localhost:6379)
//   - REDIS_DB: Redis database number (default: 15 for testing)
//
// If Redis is not available, the test will be skipped.
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()

	// Get Redis address from environment or use default
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB := 15 // Use DB 15 for testing to avoid conflicts
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		var err error
		if _, err = fmt.Sscanf(dbStr, "%d", &redisDB); err != nil {
			t.Logf("Warning: invalid REDIS_DB value, using default: %v", err)
			redisDB = 15
		}
	}

	// Create Redis client
	redisConfig := &redis.Config{
		Addr:         redisAddr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           redisDB,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	redisClient, err := redis.NewClient(redisConfig)
	if err != nil {
		t.Skipf("Redis not available for integration test: %v. Set REDIS_ADDR to run integration tests.", err)
	}

	// Test Redis connection
	if err := redisClient.Ping(ctx); err != nil {
		t.Skipf("Redis connection failed: %v", err)
	}

	// Clean up test database
	if err := redisClient.GetClient().FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: failed to flush test database: %v", err)
	}

	env := &TestEnvironment{
		Redis: redisClient,
		ctx:   ctx,
	}

	// Register cleanup
	t.Cleanup(func() {
		env.Cleanup(t)
	})

	return env
}

// Cleanup cleans up test resources.
func (e *TestEnvironment) Cleanup(t *testing.T) {
	if e.Redis != nil {
		// Flush test database
		if err := e.Redis.GetClient().FlushDB(e.ctx).Err(); err != nil {
			t.Logf("Warning: failed to flush test database during cleanup: %v", err)
		}

		// Close Redis connection
		if err := e.Redis.Close(); err != nil {
			t.Logf("Warning: failed to close Redis connection: %v", err)
		}
	}
}

// Context returns the test context.
func (e *TestEnvironment) Context() context.Context {
	return e.ctx
}
