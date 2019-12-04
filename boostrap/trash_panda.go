package boostrap

import (
	"context"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/proxy"
	"github.com/orbs-network/trash-panda/proxy/adapter/transparent"
)

func NewTrashPanda(ctx context.Context, ids ...uint32) {
	logger := config.GetLogger()
	httpConfig := httpserver.NewServerConfig("localhost:9876")
	server := httpserver.NewHttpServer(ctx, httpConfig, logger)

	for _, id := range ids {
		cfg := proxy.Config{
			VirtualChainId: id,
			Endpoints:      []string{"http://localhost:8080"},
		}

		proxy.NewService(cfg, transparent.NewTransparentAdapter(cfg), logger).UpdateRoutes(server)
	}
}
