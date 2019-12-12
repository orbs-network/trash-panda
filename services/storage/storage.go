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
	bolt "go.etcd.io/bbolt"
	"time"
)

type TxProcessor func(incomingTransactions map[string]*protocol.SignedTransaction) map[string]protocol.TransactionStatus

type Stats struct {
	Incoming  int
	Processed int
	Total     int
}

type Storage interface {
	StoreTransaction(signedTx *protocol.SignedTransaction, status protocol.TransactionStatus) error
	GetIncomingTransactions(ctx context.Context, batchSize uint) (map[string]*protocol.SignedTransaction, error)
	UpdateTransactionStatus(txId string, status protocol.TransactionStatus) error
	Stats() (Stats, error)
	Shutdown() error
}

type storage struct {
	logger log.Logger
	db     *bolt.DB
}

const INCOMING = "incoming"
const PROCESSED = "processed"
const TRANSACTIONS = "transactions"
const TRANSACTION_STATUS = "status"

var BUCKETS = []string{
	INCOMING,
	PROCESSED,
	TRANSACTIONS,
	TRANSACTION_STATUS,
}

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
	if err := storage.createBuckets(); err != nil {
		return nil, err
	}

	storage.waitForShutdown(ctx)
	return storage, nil
}

func NewStorageForChain(ctx context.Context, logger log.Logger, dbPath string, vcid uint32, readOnly bool) (Storage, error) {
	return NewStorage(ctx, logger, fmt.Sprintf("%s/vchain-%d.bolt", dbPath, vcid), readOnly)
}

// FIXME probably shouldn't roll back ever?
func (s *storage) StoreTransaction(signedTx *protocol.SignedTransaction, status protocol.TransactionStatus) error {
	return s.db.Batch(func(tx *bolt.Tx) error {
		txIdRaw := digest.CalcTxId(signedTx.Transaction())
		if err := s.storeSignedTransaction(tx, txIdRaw, signedTx); err != nil {
			return err
		}

		if err := s.updateTransactionStatus(tx, txIdRaw, status); err != nil {
			return err
		}

		if !isProcessed(status) {
			if err := s.putTransactionIntoQueue(tx, INCOMING, txIdRaw); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *storage) UpdateTransactionStatus(txId string, status protocol.TransactionStatus) error {
	return s.db.Batch(func(tx *bolt.Tx) error {
		if isProcessed(status) {
			txIdRaw, _ := encoding.DecodeHex(txId)
			if err := s.removeTransactionFromQueue(tx, INCOMING, txIdRaw); err != nil {
				return err
			}

			if err := s.putTransactionIntoQueue(tx, PROCESSED, txIdRaw); err != nil {
				return err
			}

			if err := s.updateTransactionStatus(tx, txIdRaw, status); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *storage) storeSignedTransaction(tx *bolt.Tx, txIdRaw []byte, signedTx *protocol.SignedTransaction) error {
	queueBucket := tx.Bucket([]byte(TRANSACTIONS))

	s.logger.Info("saving transaction", log.String("txId", encoding.EncodeHex(txIdRaw)))
	return queueBucket.Put(txIdRaw, signedTx.Raw())
}

func (s *storage) putTransactionIntoQueue(tx *bolt.Tx, queue string, txIdRaw []byte) error {
	queueBucket := tx.Bucket([]byte(queue))

	s.logger.Info("adding transaction to queue", log.String("queue", queue), log.String("txId", encoding.EncodeHex(txIdRaw)))
	return queueBucket.Put(txIdRaw, WriteInt64(time.Now().UnixNano()))
}

func (s *storage) removeTransactionFromQueue(tx *bolt.Tx, queue string, txIdRaw []byte) error {
	queueBucket := tx.Bucket([]byte(queue))

	s.logger.Info("removing transaction from queue", log.String("queue", queue), log.String("txId", encoding.EncodeHex(txIdRaw)))
	return queueBucket.Delete(txIdRaw)
}

func (s *storage) GetIncomingTransactions(ctx context.Context, batchSize uint) (signedTransaction map[string]*protocol.SignedTransaction, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		signedTransaction = s.readIncomingTransactionsBatch(tx, batchSize)
		return nil
	})
	return
}

func (s *storage) readIncomingTransactionsBatch(tx *bolt.Tx, batchSize uint) (incomingTransactions map[string]*protocol.SignedTransaction) {
	incomingBucket := tx.Bucket([]byte(INCOMING))
	transactionsBucket := tx.Bucket([]byte(TRANSACTIONS))

	incomingTransactions = make(map[string]*protocol.SignedTransaction)
	cursor := incomingBucket.Cursor()

	txIdRaw, _ := cursor.First()
	for i := uint(0); len(txIdRaw) != 0 && i < batchSize; i++ {
		signedTxRaw := transactionsBucket.Get(txIdRaw)
		incomingTransactions[encoding.EncodeHex(txIdRaw)] = protocol.SignedTransactionReader(signedTxRaw)
		txIdRaw, _ = cursor.Next()
	}

	return incomingTransactions
}

func (s *storage) updateTransactionStatus(tx *bolt.Tx, txId []byte, status protocol.TransactionStatus) error {
	txPool := tx.Bucket([]byte(TRANSACTION_STATUS))

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

func (s *storage) createBuckets() error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}

	for _, bucket := range BUCKETS {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}

func (s *storage) Stats() (stats Stats, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		stats.Incoming = tx.Bucket([]byte(INCOMING)).Stats().KeyN
		stats.Processed = tx.Bucket([]byte(PROCESSED)).Stats().KeyN
		stats.Total = tx.Bucket([]byte(TRANSACTIONS)).Stats().KeyN

		return nil
	})

	return
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
