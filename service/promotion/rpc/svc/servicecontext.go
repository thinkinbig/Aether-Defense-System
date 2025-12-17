// Package svc provides service context for promotion RPC service.
package svc

import (
	"fmt"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/common/redis"
	"github.com/aether-defense-system/service/promotion/rpc/config-internal"
	"github.com/aether-defense-system/service/promotion/rpc/repo"
)

// ServiceContext represents the service context for promotion RPC service.
type ServiceContext struct {
	Config     *config.Config
	DB         *database.Client
	Redis      *redis.Client
	CouponRepo *repo.CouponRepo
}

// NewServiceContext creates a new service context.
func NewServiceContext(c *config.Config) *ServiceContext {
	var dbClient *database.Client
	var couponRepo *repo.CouponRepo
	var redisClient *redis.Client

	// Initialize database client if DSN is configured
	if c.Database.DSN != "" {
		client, err := database.NewClient(&c.Database)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		dbClient = client
		couponRepo = repo.NewCouponRepo(client.DB())
	}

	// Initialize Inventory Redis client only when configured.
	//
	// In unit tests (and some lightweight deployments) we don't always have Redis available.
	// If inventoryRedis is not set in config, keep Redis nil and let business logic decide.
	if c.InventoryRedis.Addr != "" || c.InventoryRedis.Host != "" {
		var err error
		redisClient, err = redis.NewClient(&c.InventoryRedis)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize Redis: %v", err))
		}
	}

	return &ServiceContext{
		Config:     c,
		DB:         dbClient,
		Redis:      redisClient,
		CouponRepo: couponRepo,
	}
}
