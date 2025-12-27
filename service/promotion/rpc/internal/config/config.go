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
	Database       database.Config `json:"database" yaml:"database"`
	InventoryRedis redis.Config    `json:"inventoryRedis" yaml:"inventoryRedis"`
}
