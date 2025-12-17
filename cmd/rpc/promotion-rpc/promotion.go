// Package main starts the promotion RPC service.
package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/promotion/rpc"
)

var configFile = flag.String("f", "service/promotion/rpc/etc/promotion.yaml", "the config file")

// Config mirrors the RPC server configuration for promotion service.
type Config struct {
	zrpc.RpcServerConf
}

func main() {
	flag.Parse()

	var c Config
	conf.MustLoad(*configFile, &c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		rpc.RegisterPromotionServiceServer(grpcServer, &rpc.PromotionService{})
	})
	defer s.Stop()

	_, _ = fmt.Printf("Starting promotion rpc server at %s...\n", c.ListenOn)
	s.Start()
}
