package proxy

import (
	"context"
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/services/storage"
	"github.com/orbs-network/trash-panda/transport"
	"net/http"
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

func (s *service) UpdateRoutes(server *httpserver.HttpServer) {
	for _, h := range s.handlers {
		server.RegisterHttpHandler(server.Router(), s.getPath(h.Path()), true, s.wrapHandler(h))
	}
}

func (s *service) RelayTransactions(ctx context.Context) {
	sendTxHandler := s.findHandler("send-transaction")

	handle := govnr.Forever(ctx, "http server", config.NewErrorHandler(s.logger), func() {
		s.storage.ProcessIncomingTransactions(func(txId []byte, signedTransaction *protocol.SignedTransaction) (protocol.TransactionStatus, error) {
			_, output, err := sendTxHandler.Handle((&client.SendTransactionRequestBuilder{
				SignedTransaction: protocol.SignedTransactionBuilderFromRaw(signedTransaction.Raw()),
			}).Build().Raw())

			if err != nil && err.ToError() != nil {
				// FIXME error handling
				return protocol.TRANSACTION_STATUS_RESERVED, err.ToError()
			}

			s.logger.Info("relay response", log.Stringable("response", output))

			return output.(*client.SendTransactionResponse).TransactionStatus(), nil
		})

		<-time.After(100 * time.Millisecond)
	})

	supervisor := &govnr.TreeSupervisor{}
	supervisor.Supervise(handle)
}

func (s *service) storeInput(message membuffers.Message) error {
	switch message.(type) {
	case *client.SendTransactionRequest:
		return s.storage.StoreIncomingTransaction(message.(*client.SendTransactionRequest).SignedTransaction())
	}

	return nil
}

func (s *service) getPath(path string) string {
	return fmt.Sprintf("/vchains/%d%s", s.config.VirtualChainId, path)
}

func (s *service) wrapHandler(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, e := readInput(r)
		if e != nil {
			s.writeErrorResponseAndLog(w, e)
			return
		}
		input, output, err := h.Handle(bytes)

		if err := s.storeInput(input); err != nil {
			s.logger.Error("failed to store incoming transaction", log.Error(err))
		}

		if output != nil {
			s.logger.Info("responded with", log.Stringable("response", output))
			s.writeMembuffResponse(w, output, err.Code, nil)
		} else {
			s.logger.Error("error occurred", err.LogField)
			s.writeErrorResponseAndLog(w, err)
		}
	}
}

func (s *service) findHandler(name string) Handler {
	for _, h := range s.handlers {
		if h.Name() == name {
			return h
		}
	}

	return nil
}
