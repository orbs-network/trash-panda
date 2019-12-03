package proxy

import (
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
)

type Service struct {
	logger log.Logger
	config Config
}

type Config struct {
	VirtualChainId uint32
	Endpoints      []string
}

func NewService(cfg Config, logger log.Logger) *Service {
	return &Service{
		config: cfg,
		logger: logger,
	}
}

func (s *Service) getPath(path string) string {
	return fmt.Sprintf("/vchains/%d%s", s.config.VirtualChainId, path)
}

func (s *Service) UpdateRoutes(server *httpserver.HttpServer) {
	server.RegisterHttpHandler(server.Router(), s.getPath("/api/v1/send-transaction"), true, s.sendTransactionHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath("/api/v1/send-transaction-async"), true, s.sendTransactionAsyncHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath("/api/v1/run-query"), true, s.runQueryHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath("/api/v1/get-transaction-status"), true, s.getTransactionStatusHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath("/api/v1/get-transaction-receipt-proof"), true, s.getTransactionReceiptProofHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath("/api/v1/get-block"), true, s.getBlockHandler)
}
