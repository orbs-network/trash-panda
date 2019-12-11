package storage

import (
	"context"
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/orbs-client-sdk-go/crypto/digest"
	"github.com/orbs-network/orbs-client-sdk-go/crypto/encoding"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/config"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"time"
)

type TxProcessor func(incomingTransactions map[string]*protocol.SignedTransaction) map[string]protocol.TransactionStatus

type Storage interface {
	StoreIncomingTransaction(signedTx *protocol.SignedTransaction) error
	ProcessIncomingTransactions(ctx context.Context, batchSize uint, f TxProcessor) error
	Shutdown() error
}

type storage struct {
	logger log.Logger
	db     *bolt.DB
}

const INCOMING_TRANSACTIONS = "incoming"
const PROCESSED_TRANSACTIONS = "processed"
const TRANSACTION_STATUS = "status"

func NewStorage(ctx context.Context, logger log.Logger, dataSource string, readOnly bool) (Storage, error) {
	boltDb, err := bolt.Open(dataSource, 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: readOnly,
	})
	if err != nil {
		return nil, err
	}

	storage := &storage{
		logger,
		boltDb,
	}
	storage.waitForShutdown(ctx)

	return storage, nil
}

func NewStorageForChain(ctx context.Context, logger log.Logger, dbPath string, vcid uint32, readOnly bool) (Storage, error) {
	return NewStorage(ctx, logger, fmt.Sprintf("%s/vchain-%d.bolt", dbPath, vcid), readOnly)
}

func (s *storage) StoreIncomingTransaction(signedTx *protocol.SignedTransaction) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			s.logger.Error("rolling back!")
			tx.Rollback()
		}
	}()

	if err := s.storeSignedTransaction(tx, INCOMING_TRANSACTIONS, signedTx); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *storage) storeSignedTransaction(tx *bolt.Tx, pool string, signedTx *protocol.SignedTransaction) error {
	txPool, err := tx.CreateBucketIfNotExists([]byte(pool))
	if err != nil {
		return err
	}

	s.logger.Info("saving incoming transaction", log.Stringable("tx", signedTx))

	txIdRaw := digest.CalcTxId(signedTx.Transaction())
	return txPool.Put(txIdRaw, signedTx.Raw())
}

func (s *storage) ProcessIncomingTransactions(ctx context.Context, batchSize uint, f TxProcessor) error {
	if batchSize == 0 {
		return errors.New("batch size is 0")
	}

	for {
		select {
		case <-ctx.Done():
			break
		default:

		}

		tx, err := s.db.Begin(true)
		if err != nil {
			return err
		}
		txPool, err := tx.CreateBucketIfNotExists([]byte(INCOMING_TRANSACTIONS))
		if err != nil {
			return err
		}

		incomingTransactions := s.readIncomingTransactionsBatch(txPool, batchSize)

		if len(incomingTransactions) > 0 {
			for txId, status := range f(incomingTransactions) {
				if err == nil && isProcessed(status) {
					txIdRaw, _ := encoding.DecodeHex(txId)
					if err := s.storeProcessedTransaction(tx, txIdRaw, incomingTransactions[txId]); err != nil {
						return err
					}

					if err := s.storeProcessedTransactionStatus(tx, txIdRaw, status); err != nil {
						return err
					}

					txPool.Delete(txIdRaw)
				}
			}
		}

		return tx.Commit()
	}
}

func (s *storage) readIncomingTransactionsBatch(txPool *bolt.Bucket, batchSize uint) (incomingTransactions map[string]*protocol.SignedTransaction) {
	incomingTransactions = make(map[string]*protocol.SignedTransaction)
	cursor := txPool.Cursor()

	txIdRaw, signedTxRaw := cursor.First()
	for i := uint(0); len(txIdRaw) != 0 && i < batchSize; i++ {
		incomingTransactions[encoding.EncodeHex(txIdRaw)] = protocol.SignedTransactionReader(signedTxRaw)
		txIdRaw, signedTxRaw = cursor.Next()
	}

	return incomingTransactions
}

func (s *storage) storeProcessedTransaction(tx *bolt.Tx, txId []byte, signedTx *protocol.SignedTransaction) error {
	txPool, err := tx.CreateBucketIfNotExists([]byte(PROCESSED_TRANSACTIONS))
	if err != nil {
		return err
	}

	return txPool.Put(txId, signedTx.Raw())
}

func (s *storage) storeProcessedTransactionStatus(tx *bolt.Tx, txId []byte, status protocol.TransactionStatus) error {
	txPool, err := tx.CreateBucketIfNotExists([]byte(TRANSACTION_STATUS))
	if err != nil {
		return err
	}

	s.logger.Info("updating status", log.Bytes("txId", txId), log.Stringable("status", status))

	// FIXME store number
	return txPool.Put(txId, []byte(status.String()))
}

func (s *storage) Shutdown() (err error) {
	if err = s.db.Sync(); err != nil {
		s.logger.Error("failed to synchronize storage on shutdown", log.Error(err))
	}

	if err = s.db.Close(); err != nil {
		s.logger.Error("failed to close storage on shutdown", log.Error(err))
	}

	s.logger.Info("storage shut down")

	return
}

func (s *storage) waitForShutdown(ctx context.Context) {
	govnr.Once(config.NewErrorHandler(s.logger), func() {
		select {
		case <-ctx.Done():
			s.Shutdown()
		}
	})
}

func isProcessed(status protocol.TransactionStatus) bool {
	switch status {
	case protocol.TRANSACTION_STATUS_COMMITTED:
		return true
	case protocol.TRANSACTION_STATUS_DUPLICATE_TRANSACTION_ALREADY_COMMITTED:
		return true
		// FIXME deal with rejections
	}

	return false
}
