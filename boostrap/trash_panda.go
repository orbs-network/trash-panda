package boostrap

import (
	"context"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/proxy"
)

func NewTrashPanda(ctx context.Context, adapterFactory func(proxy.Config) proxy.ProxyAdapter, httpAddress string, ids ...uint32) {
	logger := config.GetLogger()
	httpConfig := httpserver.NewServerConfig(httpAddress)
	server := httpserver.NewHttpServer(ctx, httpConfig, logger)

	for _, id := range ids {
		cfg := proxy.Config{
			VirtualChainId: id,
			Endpoints:      []string{"http://localhost:8080"},
		}

		proxy.NewService(cfg, adapterFactory(cfg), logger).UpdateRoutes(server)
	}
}
