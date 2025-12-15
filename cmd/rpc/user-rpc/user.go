// Package main starts the user RPC service.
package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/user/rpc"
)

var configFile = flag.String("f", "service/user/rpc/etc/user.yaml", "the config file")

// Config mirrors the RPC server configuration for user service.
type Config struct {
	zrpc.RpcServerConf
}

func main() {
	flag.Parse()

	var c Config
	conf.MustLoad(*configFile, &c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		rpc.RegisterUserServiceServer(grpcServer, &rpc.UserService{})
	})
	defer s.Stop()

	_, _ = fmt.Printf("Starting user rpc server at %s...\n", c.ListenOn)
	s.Start()
}
