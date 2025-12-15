// Package svc provides service context for trade RPC service.
package svc

import "github.com/aether-defense-system/service/trade/rpc/internal/config"

// ServiceContext represents the service context for trade RPC service.
type ServiceContext struct {
	Config *config.Config
}

// NewServiceContext creates a new service context.
func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
