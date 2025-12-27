// Package config defines configuration for the trade HTTP API.
package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// AuthConf represents JWT authentication configuration.
type AuthConf struct {
	AccessSecret string `json:"accessSecret" yaml:"accessSecret"`
	AccessExpire int64  `json:"accessExpire" yaml:"accessExpire"`
}

// Config defines configuration for the trade HTTP API.
// Note: fieldalignment warnings for this struct are acceptable in this project
// because it is constructed infrequently and not on hot paths.
type Config struct { //nolint:govet
	rest.RestConf
	Auth     AuthConf            `json:"auth" yaml:"auth"`
	TradeRPC *zrpc.RpcClientConf `json:"tradeRpc" yaml:"tradeRpc"`
}
