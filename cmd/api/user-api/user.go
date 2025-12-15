// Package main starts the user HTTP API service.
package main

import (
	"flag"
	"fmt"

	"github.com/aether-defense-system/cmd/api/user-api/internal/config"
	"github.com/aether-defense-system/cmd/api/user-api/internal/handler"
	"github.com/aether-defense-system/cmd/api/user-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "cmd/api/user-api/etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(&c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	_, _ = fmt.Printf("Starting user-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
