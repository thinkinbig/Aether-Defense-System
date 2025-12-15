package svc

import "github.com/aether-defense-system/service/trade/rpc/internal/config"

type ServiceContext struct {
	Config *config.Config
}

func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
