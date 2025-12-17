// Package main starts the trade RPC service.
package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/trade/rpc"
)

var configFile = flag.String("f", "service/trade/rpc/etc/trade.yaml", "the config file")

// Config mirrors the RPC server configuration for trade service.
type Config struct {
	zrpc.RpcServerConf
}

func main() {
	flag.Parse()

	var c Config
	conf.MustLoad(*configFile, &c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		rpc.RegisterTradeServiceServer(grpcServer, &rpc.TradeService{})
	})
	defer s.Stop()

	_, _ = fmt.Printf("Starting trade rpc server at %s...\n", c.RpcServerConf.ListenOn)
	s.Start()
}
