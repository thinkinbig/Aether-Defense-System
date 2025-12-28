// Package rpc provides configuration structures for promotion RPC service.
package rpc

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/common/redis"
)

// Config represents the configuration for promotion RPC service.
// This is a public structure that can be used by cmd/rpc/promotion-rpc
// without importing internal packages.
type Config struct {
	zrpc.RpcServerConf
	Database database.Config `json:"database" yaml:"database"`
	// InventoryRedis is the Redis used by business logic (stock deduction, etc.).
	// zrpc.RpcServerConf already contains a Redis field (redis.RedisKeyConf) used for RPC auth,
	// so we must not reuse the same config key here.
	InventoryRedis redis.Config `json:"inventoryRedis" yaml:"inventoryRedis"`
}
