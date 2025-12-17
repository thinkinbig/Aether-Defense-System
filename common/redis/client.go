// Package redis provides Redis client functionality with Lua script support
// optimized for high-concurrency operations in the Aether Defense System.
//
// This package implements atomic operations using Lua scripts to ensure
// data consistency in high-traffic scenarios like inventory deduction.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with additional functionality for Lua scripts
// and atomic operations required by the Aether Defense System.
type Client struct {
	rdb     *redis.Client
	scripts map[string]*redis.Script
}

// Config holds Redis client configuration.
type Config struct {
	Addr         string        `json:"addr" yaml:"addr"`                     // Redis server address
	Password     string        `json:"password" yaml:"password"`             // Redis password
	DB           int           `json:"db" yaml:"db"`                         // Redis database number
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`           // Connection pool size
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"` // Minimum idle connections
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`     // Connection timeout
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`     // Read timeout
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`   // Write timeout
}

// DefaultConfig returns a default Redis configuration optimized for high concurrency.
func DefaultConfig() *Config {
	return &Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     100, // High pool size for concurrent operations
		MinIdleConns: 10,  // Keep connections warm
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewClient creates a new Redis client with the given configuration.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	client := &Client{
		rdb:     rdb,
		scripts: make(map[string]*redis.Script),
	}

	// Load predefined scripts
	client.loadPredefinedScripts()

	return client, nil
}

// loadPredefinedScripts loads commonly used Lua scripts for atomic operations.
func (c *Client) loadPredefinedScripts() {
	// Inventory deduction script (matches the existing decr_stock.lua)
	decrStockScript := `
-- KEYS[1]: Inventory Key
-- ARGV[1]: Deduction Quantity

local stock = redis.call('GET', KEYS[1])

if (stock == false) then
    return {err = "Inventory Key does not exist"}
end

if (tonumber(stock) < tonumber(ARGV[1])) then
    return {err = "Insufficient inventory"}
end

redis.call('DECRBY', KEYS[1], ARGV[1])
return {result = "Inventory deduction successful"}
`

	// Enhanced inventory deduction with user tracking
	decrStockWithUserScript := `
-- KEYS[1]: Inventory Key
-- KEYS[2]: User purchase tracking set
-- ARGV[1]: Deduction Quantity
-- ARGV[2]: User ID

-- Check if user already purchased
if redis.call('SISMEMBER', KEYS[2], ARGV[2]) == 1 then
    return {err = "User already purchased"}
end

-- Check inventory
local stock = redis.call('GET', KEYS[1])
if (stock == false) then
    return {err = "Inventory Key does not exist"}
end

if (tonumber(stock) < tonumber(ARGV[1])) then
    return {err = "Insufficient inventory"}
end

-- Atomic deduction and user tracking
redis.call('DECRBY', KEYS[1], ARGV[1])
redis.call('SADD', KEYS[2], ARGV[2])

return {result = "Inventory deduction successful"}
`

	// Set with expiration if not exists
	setNXWithExpireScript := `
-- KEYS[1]: Key
-- ARGV[1]: Value
-- ARGV[2]: Expiration in seconds

if redis.call('EXISTS', KEYS[1]) == 1 then
    return {err = "Key already exists"}
end

redis.call('SETEX', KEYS[1], ARGV[2], ARGV[1])
return {result = "Key set successfully"}
`

	// Increment with expiration
	incrWithExpireScript := `
-- KEYS[1]: Key
-- ARGV[1]: Increment value
-- ARGV[2]: Expiration in seconds

local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[2])
end

return current
`

	scripts := map[string]string{
		"decrStock":         decrStockScript,
		"decrStockWithUser": decrStockWithUserScript,
		"setNXWithExpire":   setNXWithExpireScript,
		"incrWithExpire":    incrWithExpireScript,
	}

	for name, script := range scripts {
		c.scripts[name] = redis.NewScript(script)
	}
}

// DecrStock performs atomic inventory deduction using Lua script.
// Returns error if inventory is insufficient or key doesn't exist.
func (c *Client) DecrStock(ctx context.Context, inventoryKey string, quantity int64) error {
	script, exists := c.scripts["decrStock"]
	if !exists {
		return fmt.Errorf("decrStock script not found")
	}

	result, err := script.Run(ctx, c.rdb, []string{inventoryKey}, quantity).Result()
	if err != nil {
		return fmt.Errorf("failed to execute decrStock script: %w", err)
	}

	// Parse result
	resultMap, ok := result.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("unexpected script result format")
	}

	if errMsg, hasErr := resultMap["err"]; hasErr {
		return fmt.Errorf("script error: %v", errMsg)
	}

	return nil
}

// DecrStockWithUser performs atomic inventory deduction with user tracking.
// Prevents duplicate purchases by the same user.
func (c *Client) DecrStockWithUser(ctx context.Context, inventoryKey, userSetKey string, quantity, userID int64) error {
	script, exists := c.scripts["decrStockWithUser"]
	if !exists {
		return fmt.Errorf("decrStockWithUser script not found")
	}

	result, err := script.Run(ctx, c.rdb, []string{inventoryKey, userSetKey}, quantity, userID).Result()
	if err != nil {
		return fmt.Errorf("failed to execute decrStockWithUser script: %w", err)
	}

	// Parse result
	resultMap, ok := result.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("unexpected script result format")
	}

	if errMsg, hasErr := resultMap["err"]; hasErr {
		return fmt.Errorf("script error: %v", errMsg)
	}

	return nil
}

// SetNXWithExpire sets a key only if it doesn't exist, with expiration.
// This is useful for distributed locks and idempotency tokens.
func (c *Client) SetNXWithExpire(ctx context.Context, key, value string,
	expiration time.Duration,
) error {
	script, exists := c.scripts["setNXWithExpire"]
	if !exists {
		return fmt.Errorf("setNXWithExpire script not found")
	}

	result, err := script.Run(ctx, c.rdb, []string{key}, value, int64(expiration.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to execute setNXWithExpire script: %w", err)
	}

	// Parse result
	resultMap, ok := result.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("unexpected script result format")
	}

	if errMsg, hasErr := resultMap["err"]; hasErr {
		return fmt.Errorf("script error: %v", errMsg)
	}

	return nil
}

// IncrWithExpire increments a counter and sets expiration on first increment.
// Useful for rate limiting and statistics.
func (c *Client) IncrWithExpire(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	script, exists := c.scripts["incrWithExpire"]
	if !exists {
		return 0, fmt.Errorf("incrWithExpire script not found")
	}

	result, err := script.Run(ctx, c.rdb, []string{key}, 1, int64(expiration.Seconds())).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to execute incrWithExpire script: %w", err)
	}

	count, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected script result type: %T", result)
	}

	return count, nil
}

// ExecuteScript executes a custom Lua script.
func (c *Client) ExecuteScript(ctx context.Context, script string, keys []string,
	args ...interface{},
) (interface{}, error) {
	luaScript := redis.NewScript(script)
	return luaScript.Run(ctx, c.rdb, keys, args...).Result()
}

// GetClient returns the underlying Redis client for direct operations.
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// Close closes the Redis client connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping tests the Redis connection.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Set sets a key-value pair with optional expiration.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists checks if keys exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.rdb.Expire(ctx, key, expiration).Err()
}

// HSet sets fields in a hash.
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

// HGet gets a field from a hash.
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.rdb.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from a hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// SAdd adds members to a set.
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SAdd(ctx, key, members...).Err()
}

// SIsMember checks if a member exists in a set.
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(ctx, key, member).Result()
}

// KeyNamingHelper provides standardized key naming for the system.
type KeyNamingHelper struct{}

// NewKeyNamingHelper creates a new key naming helper.
func NewKeyNamingHelper() *KeyNamingHelper {
	return &KeyNamingHelper{}
}

// InventoryKey generates a standardized inventory key.
func (k *KeyNamingHelper) InventoryKey(courseID int64) string {
	return fmt.Sprintf("promotion:stock:%d", courseID)
}

// UserPurchaseKey generates a key for tracking user purchases.
func (k *KeyNamingHelper) UserPurchaseKey(courseID int64) string {
	return fmt.Sprintf("promotion:purchased:%d", courseID)
}

// OrderLockKey generates a key for order processing locks.
func (k *KeyNamingHelper) OrderLockKey(orderID int64) string {
	return fmt.Sprintf("trade:lock:%d", orderID)
}

// UserSessionKey generates a key for user sessions.
func (k *KeyNamingHelper) UserSessionKey(userID int64) string {
	return fmt.Sprintf("user:session:%d", userID)
}

// RateLimitKey generates a key for rate limiting.
func (k *KeyNamingHelper) RateLimitKey(userID int64, action string) string {
	return fmt.Sprintf("ratelimit:%s:%d", action, userID)
}
