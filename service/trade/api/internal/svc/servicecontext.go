// Package svc wires configuration and external dependencies for trade-api.
package svc

import (
	"github.com/aether-defense-system/service/trade/api/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/tradeservice"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext wires configuration and external dependencies for trade-api.
type ServiceContext struct {
	Config    *config.Config
	TradeRPC  tradeservice.TradeService
	JWTSecret string // JWT access secret for authentication
}

// NewServiceContext creates a new ServiceContext.
func NewServiceContext(c *config.Config) *ServiceContext {
	var tradeRPC tradeservice.TradeService
	if c.TradeRPC != nil {
		tradeRPC = tradeservice.NewTradeService(zrpc.MustNewClient(*c.TradeRPC))
	}
	return &ServiceContext{
		Config:    c,
		TradeRPC:  tradeRPC,
		JWTSecret: c.Auth.AccessSecret,
	}
}
