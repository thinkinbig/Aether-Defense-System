// Package svc provides service context for trade RPC service.
package svc

import (
	"fmt"

	"github.com/zeromicro/go-zero/zrpc"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/common/mq"
	"github.com/aether-defense-system/service/promotion/rpc/promotionservice"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/repo"
	"github.com/aether-defense-system/service/user/rpc/userservice"
)

// ServiceContext represents the service context for trade RPC service.
type ServiceContext struct {
	Config       *config.Config
	DB           *database.Client
	OrderRepo    *repo.OrderRepo
	UserRPC      userservice.UserService
	PromotionRPC promotionservice.PromotionService
	RocketMQ     *mq.TransactionProducer
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

	// Initialize User RPC client
	var userRPC userservice.UserService
	if c.UserRPC.Etcd.Key != "" || len(c.UserRPC.Etcd.Hosts) > 0 {
		userClient := zrpc.MustNewClient(c.UserRPC)
		userRPC = userservice.NewUserService(userClient)
	}

	// Initialize Promotion RPC client
	var promotionRPC promotionservice.PromotionService
	if c.PromotionRPC.Etcd.Key != "" || len(c.PromotionRPC.Etcd.Hosts) > 0 {
		promotionClient := zrpc.MustNewClient(c.PromotionRPC)
		promotionRPC = promotionservice.NewPromotionService(promotionClient)
	}

	// RocketMQ producer will be initialized lazily when needed (in PlaceOrder logic)
	// to allow for dependency injection of transaction executors

	return &ServiceContext{
		Config:       c,
		DB:           dbClient,
		OrderRepo:    orderRepo,
		UserRPC:      userRPC,
		PromotionRPC: promotionRPC,
		RocketMQ:     nil, // Will be set when transaction producer is created
	}
}
