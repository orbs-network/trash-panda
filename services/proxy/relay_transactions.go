package proxy

import (
	"context"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/config"
	"time"
)

func (s *service) RelayTransactions(ctx context.Context) {
	sendTxHandler := s.findHandler("send-transaction")

	handle := govnr.Forever(ctx, "http server", config.NewErrorHandler(s.logger), func() {
		err := s.storage.ProcessIncomingTransactions(ctx, s.config.RelayBatchSize, func(signedTransactions map[string]*protocol.SignedTransaction) map[string]protocol.TransactionStatus {
			processedTransactions := make(map[string]protocol.TransactionStatus)

			s.logger.Info("relaying transactions", log.Int("batch", len(signedTransactions)))

			for txId, signedTransaction := range signedTransactions {
				_, output, err := sendTxHandler.Handle((&client.SendTransactionRequestBuilder{
					SignedTransaction: protocol.SignedTransactionBuilderFromRaw(signedTransaction.Raw()),
				}).Build().Raw())

				if err != nil && err.ToError() != nil {
					s.logger.Info("relay error", log.Error(err.ToError()), log.String("txId", txId))
					processedTransactions[txId] = protocol.TRANSACTION_STATUS_RESERVED
				} else {
					s.logger.Info("relay response", log.Stringable("response", output))
					processedTransactions[txId] = output.(*client.SendTransactionResponse).TransactionStatus()
				}
			}

			return processedTransactions
		})

		if err != nil {
			s.logger.Error("could not relay transaction batch", log.Error(err))
		}

		<-time.After(s.config.RelayInterval)
	})

	supervisor := &govnr.TreeSupervisor{}
	supervisor.Supervise(handle)
}
