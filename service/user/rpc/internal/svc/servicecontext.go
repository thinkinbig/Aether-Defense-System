// Package svc provides service context for user RPC service.
package svc

import "github.com/aether-defense-system/service/user/rpc/internal/config"

// ServiceContext represents the service context for user RPC service.
type ServiceContext struct {
	Config *config.Config
}

// NewServiceContext creates a new service context.
func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
