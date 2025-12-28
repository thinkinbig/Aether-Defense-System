// Package config defines configuration for the user HTTP API.
package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config defines configuration for the user HTTP API.
// Note: fieldalignment warnings for this struct are acceptable in this project
// because it is constructed infrequently and not on hot paths.
type Config struct { //nolint:govet
	rest.RestConf
	UserRPC *zrpc.RpcClientConf
}
