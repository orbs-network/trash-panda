package storage

import (
	"context"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/config"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
	"time"
)

const GAMMA_ENDPOINT = "http://localhost:8080"
const GAMMA_VCHAIN = uint32(42)

func removeDB() {
	os.RemoveAll("./vchain-42.bolt")
}

func TestStorage_StoreIncomingTransaction(t *testing.T) {
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
	err = s.StoreIncomingTransaction(signedTx)
	require.NoError(t, err)
}

func TestStorage_ProcessIncomingTransactions(t *testing.T) {
	removeDB()

	account, _ := orbs.CreateAccount()
	orbsClient := orbs.NewClient(GAMMA_ENDPOINT, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	tx, txId, err := orbsClient.CreateTransaction(account.PublicKey, account.PrivateKey,
		"Music1974", "getAlbum", "Diamond Dogs")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := NewStorageForChain(ctx, config.GetLogger(), "./", GAMMA_VCHAIN, false)
	require.NoError(t, err)

	signedTx := client.SendTransactionRequestReader(tx).SignedTransaction()
	err = s.StoreIncomingTransaction(signedTx)
	require.NoError(t, err)

	transactionsProcessed := 0
	err = s.ProcessIncomingTransactions(context.Background(), 1, func(incomingTransactions map[string]*protocol.SignedTransaction) (results map[string]protocol.TransactionStatus) {
		results = make(map[string]protocol.TransactionStatus)
		for incomingTxId, _ := range incomingTransactions {
			require.EqualValues(t, txId, incomingTxId)
			results[incomingTxId] = protocol.TRANSACTION_STATUS_COMMITTED
			transactionsProcessed++
		}

		return
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, transactionsProcessed)

	transactionsProcessedTheSecondTime := 0
	err = s.ProcessIncomingTransactions(context.Background(), 1, func(incomingTransactions map[string]*protocol.SignedTransaction) (results map[string]protocol.TransactionStatus) {
		for incomingTxId, _ := range incomingTransactions {
			transactionsProcessedTheSecondTime++
			results[incomingTxId] = protocol.TRANSACTION_STATUS_COMMITTED
		}

		return
	})
	require.NoError(t, err)
	require.EqualValues(t, 0, transactionsProcessedTheSecondTime)
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
	err = s.StoreIncomingTransaction(signedTx)
	require.NoError(t, err)

	go func() {
		for {
			err = s.ProcessIncomingTransactions(context.Background(), 1, func(incomingTransactions map[string]*protocol.SignedTransaction) (results map[string]protocol.TransactionStatus) {
				fmt.Printf("processing %v", incomingTransactions)
				return
			})
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
