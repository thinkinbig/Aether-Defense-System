// Package main starts the trade RPC service.
package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/server"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"
)

var configFile = flag.String("f", "service/trade/rpc/etc/trade.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// Create service context with all dependencies
	ctx := svc.NewServiceContext(&c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		rpc.RegisterTradeServiceServer(grpcServer, server.NewTradeServiceServer(ctx))
	})
	defer s.Stop()

	_, _ = fmt.Printf("Starting trade rpc server at %s...\n", c.ListenOn)
	s.Start()
}
