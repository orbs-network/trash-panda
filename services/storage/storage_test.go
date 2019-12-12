package storage

import (
	"context"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/config"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

const GAMMA_ENDPOINT = "http://localhost:8080"
const GAMMA_VCHAIN = uint32(42)

func removeDB() {
	os.RemoveAll("./vchain-42.bolt")
}

func TestStorage_StoreTransaction(t *testing.T) {
	removeDB()

	account, _ := orbs.CreateAccount()
	orbsClient := orbs.NewClient(GAMMA_ENDPOINT, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	tx, _, err := orbsClient.CreateTransaction(account.PublicKey, account.PrivateKey,
		"Music1974", "getAlbum", "Diamond Dogs")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := NewStorageForChain(ctx, config.GetLogger(), "./", GAMMA_VCHAIN, false)
	require.NoError(t, err)

	signedTx := client.SendTransactionRequestReader(tx).SignedTransaction()
	err = s.StoreTransaction(signedTx, protocol.TRANSACTION_STATUS_RESERVED)
	require.NoError(t, err)
}

func TestStorage_GetIncomingTransactions(t *testing.T) {
	removeDB()

	account, _ := orbs.CreateAccount()
	orbsClient := orbs.NewClient(GAMMA_ENDPOINT, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	tx, txId, err := orbsClient.CreateTransaction(account.PublicKey, account.PrivateKey,
		"Music1974", "getAlbum", "Diamond Dogs")
	require.NoError(t, err)

	anotherTx, anotherTxId, _ := orbsClient.CreateTransaction(account.PublicKey, account.PrivateKey,
		"Music1974", "getAlbum", "Station to Station")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := NewStorageForChain(ctx, config.GetLogger(), "./", GAMMA_VCHAIN, false)
	require.NoError(t, err)

	signedTx := client.SendTransactionRequestReader(tx).SignedTransaction()
	err = s.StoreTransaction(signedTx, protocol.TRANSACTION_STATUS_RESERVED)
	require.NoError(t, err)

	anotherSignedTx := client.SendTransactionRequestReader(anotherTx).SignedTransaction()
	err = s.StoreTransaction(anotherSignedTx, protocol.TRANSACTION_STATUS_RESERVED)
	require.NoError(t, err)

	incomingTransactions, err := s.GetIncomingTransactions(context.Background(), 5)
	require.NoError(t, err)
	require.EqualValues(t, 2, len(incomingTransactions))

	require.EqualValues(t, signedTx.Raw(), incomingTransactions[txId].Raw())
	require.EqualValues(t, anotherSignedTx.Raw(), incomingTransactions[anotherTxId].Raw())

	s.UpdateTransactionStatus(txId, protocol.TRANSACTION_STATUS_COMMITTED)
	s.UpdateTransactionStatus(anotherTxId, protocol.TRANSACTION_STATUS_DUPLICATE_TRANSACTION_ALREADY_COMMITTED)

	incomingTransactions, err = s.GetIncomingTransactions(context.Background(), 5)
	require.NoError(t, err)
	require.EqualValues(t, 0, len(incomingTransactions))
}

func TestStorage_WaitForShutdown(t *testing.T) {
	removeDB()

	ctx, cancel := context.WithCancel(context.Background())
	s, err := NewStorageForChain(ctx, config.GetLogger(), "./", GAMMA_VCHAIN, false)
	require.NoError(t, err)

	account, _ := orbs.CreateAccount()
	orbsClient := orbs.NewClient(GAMMA_ENDPOINT, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	tx, _, err := orbsClient.CreateTransaction(account.PublicKey, account.PrivateKey,
		"Music1974", "getAlbum", "Diamond Dogs")
	require.NoError(t, err)

	signedTx := client.SendTransactionRequestReader(tx).SignedTransaction()
	err = s.StoreTransaction(signedTx, protocol.TRANSACTION_STATUS_RESERVED)
	require.NoError(t, err)

	go func() {
		for {
			txs, _ := s.GetIncomingTransactions(context.Background(), 1)
			for txId, _ := range txs {
				s.UpdateTransactionStatus(txId, protocol.TRANSACTION_STATUS_COMMITTED)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	time.Sleep(1 * time.Second)

	cancel()

	boltDb, err := bolt.Open("./vchain-42.bolt", 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: true,
	})
	require.NoError(t, err)
	boltDb.Close()
}

const MAX_TX = 100
const BATCH_SIZE = 5
const INSERT_INTEVAL = 1 * time.Millisecond
const UPDATE_INTERVAL = 2 * time.Millisecond

func Test_Concurrency(t *testing.T) {
	removeDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, err := NewStorageForChain(ctx, config.GetLogger(), "./", GAMMA_VCHAIN, false)
	require.NoError(t, err)

	account, _ := orbs.CreateAccount()
	orbsClient := orbs.NewClient(GAMMA_ENDPOINT, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	txProcessed := uint32(0)
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:

			}
			txs, _ := s.GetIncomingTransactions(context.Background(), BATCH_SIZE)
			for txId, _ := range txs {
				err := s.UpdateTransactionStatus(txId, protocol.TRANSACTION_STATUS_COMMITTED)
				require.NoError(t, err)
				txProcessed++
				atomic.AddUint32(&txProcessed, 1)
			}

			<-time.After(UPDATE_INTERVAL)
		}
	}()

	go func() {
		for i := 0; i < MAX_TX; i++ {
			select {
			case <-ctx.Done():
				break
			default:

			}

			tx, _, err := orbsClient.CreateTransaction(account.PublicKey, account.PrivateKey,
				"Music1974", "getAlbum", "Diamond Dogs", uint32(i))
			require.NoError(t, err)

			signedTx := client.SendTransactionRequestReader(tx).SignedTransaction()
			err = s.StoreTransaction(signedTx, protocol.TRANSACTION_STATUS_RESERVED)
			require.NoError(t, err)

			<-time.After(INSERT_INTEVAL)
		}
	}()

	require.Eventually(t, func() bool {
		return txProcessed == MAX_TX
	}, 2*time.Second, INSERT_INTEVAL)

	require.Eventually(t, func() bool {
		stats, err := s.Stats()
		if err != nil {
			return false
		}

		return stats.Incoming == 0 && stats.Total == MAX_TX && stats.Total == stats.Processed
	}, 2*time.Second, INSERT_INTEVAL)

}
