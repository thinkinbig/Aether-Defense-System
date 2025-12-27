// Package svc wires configuration and external dependencies for user-api.
package svc

import (
	"github.com/aether-defense-system/service/user/rpc/userservice"

	"github.com/aether-defense-system/service/user/api/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext wires configuration and external dependencies for user-api.
type ServiceContext struct {
	Config  *config.Config
	UserRPC userservice.UserService
}

// NewServiceContext creates a new ServiceContext.
func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRPC: userservice.NewUserService(zrpc.MustNewClient(*c.UserRPC)),
	}
}
