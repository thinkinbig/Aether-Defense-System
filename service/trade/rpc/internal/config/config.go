// Package config contains configuration for trade RPC service.
package config

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/common/mq"
)

// Config represents the configuration for trade RPC service.
type Config struct {
	zrpc.RpcServerConf
	RocketMQ     mq.Config          `json:"rocketmq" yaml:"rocketmq"`
	UserRPC      zrpc.RpcClientConf `json:"userRpc" yaml:"userRpc"`
	PromotionRPC zrpc.RpcClientConf `json:"promotionRpc" yaml:"promotionRpc"`
	Database     database.Config    `json:"database" yaml:"database"`
}
