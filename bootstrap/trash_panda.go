package bootstrap

import (
	"context"
	"fmt"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/services/proxy"
	"github.com/orbs-network/trash-panda/services/storage"
	"github.com/orbs-network/trash-panda/transport"
	"time"
)

func NewTrashPanda(ctx context.Context, transport transport.Transport, cfg *config.Config) *httpserver.HttpServer {
	logger := config.GetLogger()
	httpConfig := httpserver.NewServerConfig(cfg.HttpAddress)
	server := httpserver.NewHttpServer(ctx, httpConfig, logger)

	for _, vcid := range cfg.VirtualChains {
		proxyConfig := buildProxyConfig(cfg, vcid)

		s, err := storage.NewStorageForChain(ctx, logger, cfg.Database, vcid, false)
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

func buildProxyConfig(cfg *config.Config, vcid uint32) proxy.Config {
	endpoints := cfg.Endpoints

	if !cfg.Gamma {
		for i, endpoint := range endpoints {
			endpoints[i] = fmt.Sprintf("%s/vchains/%d", endpoint, vcid)
		}
	}

	return proxy.Config{
		VirtualChainId: vcid,
		Endpoints:      endpoints,
		RelayBatchSize: cfg.RelayBatchSize,
		RelayInterval:  time.Duration(cfg.RelayIntervalMs) * time.Millisecond,
	}
}
