package bootstrap

import (
	"context"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/proxy"
	"github.com/orbs-network/trash-panda/transport"
)

func NewTrashPanda(ctx context.Context, transport transport.Transport, httpAddress string, ids ...uint32) *httpserver.HttpServer {
	logger := config.GetLogger()
	httpConfig := httpserver.NewServerConfig(httpAddress)
	server := httpserver.NewHttpServer(ctx, httpConfig, logger)

	for _, id := range ids {
		cfg := proxy.Config{
			VirtualChainId: id,
			Endpoints:      []string{"http://localhost:8080"},
		}

		p := proxy.NewService(cfg, transport, logger)
		p.UpdateRoutes(server)
		p.ResendTxQueue(ctx)
	}

	return server
}
