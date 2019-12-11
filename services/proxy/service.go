package proxy

import (
	"context"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
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

func (s *service) RelayTransactions(ctx context.Context) {
	sendTxHandler := s.findHandler("send-transaction")

	handle := govnr.Forever(ctx, "http server", config.NewErrorHandler(s.logger), func() {
		s.storage.ProcessIncomingTransactions(ctx, func(txId []byte, signedTransaction *protocol.SignedTransaction) (protocol.TransactionStatus, error) {
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
