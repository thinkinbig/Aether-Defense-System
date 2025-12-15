// Package config contains configuration for user RPC service.
package config

import "github.com/zeromicro/go-zero/zrpc"

// Config represents the configuration for user RPC service.
type Config struct {
	zrpc.RpcServerConf
}
