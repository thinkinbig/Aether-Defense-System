// Package svc provides service context for user RPC service.
package svc

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/service/user/rpc/internal/config"
	"github.com/aether-defense-system/service/user/rpc/internal/repo"
)

// UserRepository defines the minimal persistence operations required by user logic.
// This makes user logic unit-testable without a real database.
type UserRepository interface {
	GetByID(ctx context.Context, userID int64) (*database.User, error)
}

// ServiceContext represents the service context for user RPC service.
type ServiceContext struct {
	Config   *config.Config
	DB       *database.Client
	UserRepo UserRepository
}

// NewServiceContext creates a new service context.
func NewServiceContext(c *config.Config) *ServiceContext {
	var dbClient *database.Client
	var userRepo UserRepository

	// Initialize database client if DSN is configured
	if c.Database.DSN != "" {
		client, err := database.NewClient(&c.Database)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		dbClient = client
		userRepo = repo.NewUserRepo(client.DB())
	}

	return &ServiceContext{
		Config:   c,
		DB:       dbClient,
		UserRepo: userRepo,
	}
}
