// Package config contains configuration for promotion RPC service.
package config

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/aether-defense-system/common/database"
)

// Config represents the configuration for promotion RPC service.
type Config struct {
	zrpc.RpcServerConf
	Database database.Config `json:"database" yaml:"database"`
}
