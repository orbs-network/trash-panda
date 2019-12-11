package storage

import (
	"context"
	"encoding/hex"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/config"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"strings"
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

	s, err := NewStorageForChain(config.GetLogger(), "./", GAMMA_VCHAIN, false)
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

	s, err := NewStorageForChain(config.GetLogger(), "./", GAMMA_VCHAIN, false)
	require.NoError(t, err)

	signedTx := client.SendTransactionRequestReader(tx).SignedTransaction()
	err = s.StoreIncomingTransaction(signedTx)
	require.NoError(t, err)

	transactionsProcessed := 0
	err = s.ProcessIncomingTransactions(context.Background(), func(txIdRaw []byte, incomingTransaction *protocol.SignedTransaction) (protocol.TransactionStatus, error) {
		require.EqualValues(t, strings.ToLower(txId), "0x"+hex.EncodeToString(txIdRaw))
		transactionsProcessed++
		return protocol.TRANSACTION_STATUS_COMMITTED, nil
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, transactionsProcessed)

	transactionsProcessedTheSecondTime := 0
	err = s.ProcessIncomingTransactions(context.Background(), func(txIdRaw []byte, incomingTransaction *protocol.SignedTransaction) (protocol.TransactionStatus, error) {
		transactionsProcessedTheSecondTime++
		return protocol.TRANSACTION_STATUS_COMMITTED, nil
	})
	require.NoError(t, err)
	require.EqualValues(t, 0, transactionsProcessedTheSecondTime)
}

func TestStorage_WaitForShutdown(t *testing.T) {
	removeDB()

	ctx, cancel := context.WithCancel(context.Background())
	s, err := NewStorageForChain(config.GetLogger(), "./", GAMMA_VCHAIN, false)
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
			s.ProcessIncomingTransactions(ctx, func(txId []byte, incomingTransaction *protocol.SignedTransaction) (protocol.TransactionStatus, error) {
				println("processing", incomingTransaction.String())
				return protocol.TRANSACTION_STATUS_RESERVED, nil
			})

			time.Sleep(100 * time.Millisecond)
		}
	}()

	s.WaitForShutdown(ctx)
	time.Sleep(1 * time.Second)

	cancel()

	boltDb, err := bolt.Open("./vchain-42.bolt", 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: true,
	})
	require.NoError(t, err)
	boltDb.Close()
}
