// Package config defines configuration for the trade HTTP API.
package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config defines configuration for the trade HTTP API.
type Config struct {
	rest.RestConf
	TradeRPC *zrpc.RpcClientConf `json:"tradeRpc" yaml:"tradeRpc"`
}
