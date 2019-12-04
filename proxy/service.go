package proxy

import (
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
)

const (
	SEND_TRANSACTION              = "/api/v1/send-transaction"
	SEND_TRANSACTION_ASYNC        = "/api/v1/send-transaction-async"
	RUN_QUERY                     = "/api/v1/run-query"
	GET_TRANSACTION_STATUS        = "/api/v1/get-transaction-status"
	GET_TRANSACTION_RECEIPT_PROOF = "/api/v1/get-transaction-receipt-proof"
	GET_BLOCK                     = "/api/v1/get-block"
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
	server.RegisterHttpHandler(server.Router(), s.getPath(SEND_TRANSACTION), true, s.sendTransactionHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(SEND_TRANSACTION_ASYNC), true, s.sendTransactionAsyncHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(RUN_QUERY), true, s.runQueryHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_STATUS), true, s.getTransactionStatusHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_RECEIPT_PROOF), true, s.getTransactionReceiptProofHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(GET_BLOCK), true, s.getBlockHandler)
}
