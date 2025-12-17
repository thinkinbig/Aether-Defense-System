// Package config contains configuration for promotion RPC service.
package config

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/common/redis"
)

// Config represents the configuration for promotion RPC service.
// This is the internal config structure used by internal packages.
// The public Config is defined in the parent rpc package.
type Config struct {
	zrpc.RpcServerConf
	Database database.Config `json:"database" yaml:"database"`
	// InventoryRedis is the Redis used by business logic (stock deduction, etc.).
	//
	// Note: zrpc.RpcServerConf already has a field named Redis (redis.RedisKeyConf) used for RPC auth.
	// Keeping our business Redis under a different key avoids config load conflicts.
	InventoryRedis redis.Config `json:"inventoryRedis" yaml:"inventoryRedis"`
}
