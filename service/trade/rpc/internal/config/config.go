// Package config contains configuration for trade RPC service.
package config

import "github.com/zeromicro/go-zero/zrpc"

// Config represents the configuration for trade RPC service.
type Config struct {
	zrpc.RpcServerConf
}
