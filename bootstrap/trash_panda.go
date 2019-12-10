package bootstrap

import (
	"context"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/services/proxy"
	"github.com/orbs-network/trash-panda/services/storage"
	"github.com/orbs-network/trash-panda/transport"
)

func NewTrashPanda(ctx context.Context, transport transport.Transport, cfg *config.Config) *httpserver.HttpServer {
	logger := config.GetLogger()
	httpConfig := httpserver.NewServerConfig(cfg.HttpAddress)
	server := httpserver.NewHttpServer(ctx, httpConfig, logger)

	for _, vcid := range cfg.VirtualChains {
		proxyConfig := proxy.Config{
			VirtualChainId: vcid,
			Endpoints:      cfg.Endpoints,
		}

		s, err := storage.NewStorageForChain(logger, "./", vcid, false)
		if err != nil {
			// FIXME error handling
			panic(err)
		}

		p := proxy.NewService(proxyConfig, s, transport, logger)
		p.UpdateRoutes(server)
		p.RelayTransactions(ctx)
	}

	return server
}
