// Package main starts the trade HTTP API service.
package main

import (
	"flag"
	"fmt"

	"github.com/aether-defense-system/service/trade/api/internal/config"
	"github.com/aether-defense-system/service/trade/api/internal/handler"
	"github.com/aether-defense-system/service/trade/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "service/trade/api/etc/trade-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(&c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	_, _ = fmt.Printf("Starting trade-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
