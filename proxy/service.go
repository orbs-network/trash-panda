package proxy

import (
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
	"net/http"
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
	logger  log.Logger
	config  Config
	adapter ProxyAdapter
}

type Config struct {
	VirtualChainId uint32
	Endpoints      []string
}

func NewService(cfg Config, adapter ProxyAdapter, logger log.Logger) *Service {
	return &Service{
		config:  cfg,
		logger:  logger,
		adapter: adapter,
	}
}

func (s *Service) getPath(path string) string {
	return fmt.Sprintf("/vchains/%d%s", s.config.VirtualChainId, path)
}

func (s *Service) UpdateRoutes(server *httpserver.HttpServer) {
	server.RegisterHttpHandler(server.Router(), s.getPath(SEND_TRANSACTION), true, s.sendTransactionHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(SEND_TRANSACTION_ASYNC), true, s.sendTransactionAsyncHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(RUN_QUERY), true, s.runQueryHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_STATUS), true, s.getTransactionStatusHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_RECEIPT_PROOF), true, s.getTransactionReceiptProofHandler)
	server.RegisterHttpHandler(server.Router(), s.getPath(GET_BLOCK), true, s.getBlockHandler)

	for _, h := range s.adapter.Handlers() {
		server.RegisterHttpHandler(server.Router(), s.getPath(h.Path()), true, s.wrapHandler(h.Handler()))
	}
}

func (s *Service) wrapHandler(handlerBuilder HandlerBuilderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, e := readInput(r)
		if e != nil {
			s.writeErrorResponseAndLog(w, e)
			return
		}
		result, err := handlerBuilder(bytes)
		if err != nil {
			s.writeErrorResponseAndLog(w, err)
		} else {
			s.writeMembuffResponse(w, result, 200, nil)
		}
	}
}
