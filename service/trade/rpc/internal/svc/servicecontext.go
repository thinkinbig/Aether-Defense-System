// Package svc provides service context for trade RPC service.
package svc

import (
	"fmt"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/repo"
)

// ServiceContext represents the service context for trade RPC service.
type ServiceContext struct {
	Config    *config.Config
	DB        *database.Client
	OrderRepo *repo.OrderRepo
}

// NewServiceContext creates a new service context.
func NewServiceContext(c *config.Config) *ServiceContext {
	var dbClient *database.Client
	var orderRepo *repo.OrderRepo

	// Initialize database client if DSN is configured
	if c.Database.DSN != "" {
		client, err := database.NewClient(&c.Database)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		dbClient = client
		orderRepo = repo.NewOrderRepo(client.DB())
	}

	return &ServiceContext{
		Config:    c,
		DB:        dbClient,
		OrderRepo: orderRepo,
	}
}
