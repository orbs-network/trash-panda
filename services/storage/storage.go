package storage

import (
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/crypto/digest"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/scribe/log"
	bolt "go.etcd.io/bbolt"
	"time"
)

type TxProcessor func(txId []byte, incomingTransaction *protocol.SignedTransaction) (protocol.TransactionStatus, error)

type Storage interface {
	StoreIncomingTransaction(signedTx *protocol.SignedTransaction) error
	ProcessIncomingTransactions(f TxProcessor) error
	Shutdown() error
}

type storage struct {
	logger log.Logger
	db     *bolt.DB
}

const INCOMING_TRANSACTIONS = "incoming"
const PROCESSED_TRANSACTIONS = "processed"
const TRANSACTION_STATUS = "status"

func NewStorage(logger log.Logger, dataSource string, readOnly bool) (Storage, error) {
	boltDb, err := bolt.Open(dataSource, 0600, &bolt.Options{
		Timeout:  5 * time.Second,
		ReadOnly: readOnly,
	})
	if err != nil {
		return nil, err
	}

	return &storage{
		logger,
		boltDb,
	}, nil
}

func NewStorageForChain(logger log.Logger, dbPath string, vcid uint32, readOnly bool) (Storage, error) {
	return NewStorage(logger, fmt.Sprintf("%s/vchain-%d.bolt", dbPath, vcid), readOnly)
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

	s.logger.Info("Storing tx", log.Stringable("tx", signedTx))

	txIdRaw := digest.CalcTxId(signedTx.Transaction())
	return txPool.Put(txIdRaw, signedTx.Raw())
}

func (s *storage) ProcessIncomingTransactions(f TxProcessor) error {
	for {
		tx, err := s.db.Begin(true)
		if err != nil {
			return err
		}

		txPool, err := tx.CreateBucketIfNotExists([]byte(INCOMING_TRANSACTIONS))
		if err != nil {
			return err
		}

		i := txPool.Cursor()

		// FIXME read in batches
		txIdRaw, signedTxRaw := i.First()
		for ; len(txIdRaw) != 0; txIdRaw, signedTxRaw = i.Next() {
			signedTx := protocol.SignedTransactionReader(signedTxRaw)

			if status, err := f(txIdRaw, signedTx); err == nil && isProcessed(status) {
				if err := s.storeProcessedTransaction(tx, txIdRaw, signedTx); err != nil {
					return err
				}

				if err := s.storeProcessedTransactionStatus(tx, txIdRaw, status); err != nil {
					return err
				}

				txPool.Delete(txIdRaw)
			}
		}

		return tx.Commit()
	}
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
	// FIXME proper shutdown
	if err = s.db.Sync(); err != nil {
		s.logger.Error("failed to synchronize storage on shutdown")
	}

	if err = s.db.Close(); err != nil {
		s.logger.Error("failed to close storage on shutdown")
	}

	s.logger.Info("storage shut down")

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
