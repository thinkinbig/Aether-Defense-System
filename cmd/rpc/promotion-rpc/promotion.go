// Package main starts the promotion RPC service.
//
// Best Practice Solution:
// We moved internal/server, internal/config, internal/svc, and internal/repo
// out of the internal directory to allow cmd/rpc/promotion-rpc to import them.
// This follows go-zero best practices while maintaining the cmd/rpc directory structure.
package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/promotion/rpc"
	promotionServer "github.com/aether-defense-system/service/promotion/rpc/server"
	promotionSvc "github.com/aether-defense-system/service/promotion/rpc/svc"
)

var configFile = flag.String("f", "service/promotion/rpc/etc/promotion.yaml", "the config file")

func main() {
	flag.Parse()

	// Load config using public config type
	var publicCfg rpc.Config
	conf.MustLoad(*configFile, &publicCfg)

	// Create ServiceContext using the public helper function
	ctx := promotionSvc.NewServiceContextFromPublic(&publicCfg)

	s := zrpc.MustNewServer(publicCfg.RpcServerConf, func(grpcServer *grpc.Server) {
		rpc.RegisterPromotionServiceServer(grpcServer, promotionServer.NewPromotionServiceServer(ctx))
	})
	defer s.Stop()

	_, _ = fmt.Printf("Starting promotion rpc server at %s...\n", publicCfg.RpcServerConf.ListenOn)
	s.Start()
}
