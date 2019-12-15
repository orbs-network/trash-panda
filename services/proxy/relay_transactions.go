package proxy

import (
	"context"
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/config"
	"sync/atomic"
	"time"
)

type statusContainer struct {
	txId   string
	status protocol.TransactionStatus
}

func (s *service) RelayTransactions(ctx context.Context) {
	sendTxHandler := s.findHandler("send-transaction")
	attempts := uint32(0)

	handle := govnr.Forever(ctx, "http server", config.NewErrorHandler(s.logger), func() {
		signedTransactions, err := s.storage.GetIncomingTransactions(ctx, s.config.RelayBatchSize)
		if err != nil {
			s.logger.Error("could not get incoming transactions batch", log.Error(err))
			<-time.After(s.config.RelayInterval)
			return
		}

		numTx := len(signedTransactions)

		s.logger.Info("relaying transactions", log.Int("batch", numTx))
		statusCH := make(chan *statusContainer)
		for txId, signedTransaction := range signedTransactions {
			txLogger := s.logger.WithTags(log.String("txId", txId))
			govnr.Once(config.NewErrorHandler(txLogger), processTx(txId, signedTransaction, statusCH, sendTxHandler, txLogger))
		}

		for i := uint(0); i < uint(numTx); i++ {
			container := <-statusCH
			if container != nil {
				if err := s.storage.UpdateTransactionStatus(container.txId, container.status); err != nil {
					s.logger.Error("could not update transaction status", log.Error(err), log.String("txId", container.txId))
				}
				atomic.AddUint32(&attempts, 1)
			}
		}
		close(statusCH)

		stats, _ := s.storage.Stats()
		println(fmt.Sprintf("%+v", stats), attempts)
		s.logger.Info("relay tick stats",
			log.Int("incoming", stats.Incoming),
			log.Int("processed", stats.Processed),
			log.Int("tx-total", stats.Total),
			log.Uint32("attempts-total", attempts))

		if err != nil {
			s.logger.Error("could not relay transaction batch", log.Error(err))
		}

		<-time.After(s.config.RelayInterval)
	})

	supervisor := &govnr.TreeSupervisor{}
	supervisor.Supervise(handle)
}

func processTx(txId string, signedTransaction *protocol.SignedTransaction, statusCH chan *statusContainer, sendTxHandler Handler, logger log.Logger) func() {
	return func() {
		_, output, err := sendTxHandler.Handle((&client.SendTransactionRequestBuilder{
			SignedTransaction: protocol.SignedTransactionBuilderFromRaw(signedTransaction.Raw()),
		}).Build().Raw())

		if err != nil && err.ToError() != nil {
			logger.Info("relay error", log.Error(err.ToError()))
			statusCH <- &statusContainer{
				txId:   txId,
				status: protocol.TRANSACTION_STATUS_RESERVED,
			}
		} else {
			logger.Info("relay response", log.Stringable("response", output))
			statusCH <- &statusContainer{
				txId:   txId,
				status: output.(*client.SendTransactionResponse).TransactionStatus(),
			}
		}
	}
}
