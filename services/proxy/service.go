package proxy

import (
	"context"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/services/storage"
	"github.com/orbs-network/trash-panda/transport"
	"time"
)

type service struct {
	logger   log.Logger
	config   Config
	storage  storage.Storage
	handlers []Handler
}

type Config struct {
	VirtualChainId uint32
	Endpoints      []string

	RelayInterval  time.Duration
	RelayBatchSize uint
}

type Proxy interface {
	UpdateRoutes(server *httpserver.HttpServer)
	RelayTransactions(ctx context.Context)
}

func NewService(cfg Config, storage storage.Storage, transport transport.Transport, logger log.Logger) *service {
	return &service{
		config:   cfg,
		logger:   logger,
		handlers: GetHandlers(cfg, transport, logger),
		storage:  storage,
	}
}
