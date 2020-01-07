package e2e

import (
	"context"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/transport"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_RunQuery(t *testing.T) {
	contractTest(t, func(t *testing.T, endpoint string, vcid uint32) {
		account, _ := orbs.CreateAccount()
		client := orbs.NewClient(endpoint, vcid, codec.NETWORK_TYPE_TEST_NET)

		query, err := client.CreateQuery(account.PublicKey, "_Info", "isAlive")
		require.NoError(t, err)

		res, err := client.SendQuery(query)
		require.NoError(t, err)
		require.EqualValues(t, res.RequestStatus, codec.REQUEST_STATUS_COMPLETED)
		require.GreaterOrEqual(t, res.BlockHeight, uint64(1))
	})

}

func Test_SendTransaction(t *testing.T) {
	contractTest(t, func(t *testing.T, endpoint string, vcid uint32) {
		account, _ := orbs.CreateAccount()
		client := orbs.NewClient(endpoint, vcid, codec.NETWORK_TYPE_TEST_NET)

		contractName := deployIncrementContractToGamma(t)

		tx, txId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc", uint32(0))
		require.NoError(t, err)

		res, err := client.SendTransaction(tx)
		require.NoError(t, err)
		require.EqualValues(t, res.ExecutionResult, codec.EXECUTION_RESULT_SUCCESS)
		require.EqualValues(t, res.RequestStatus, codec.REQUEST_STATUS_COMPLETED)
		require.EqualValues(t, uint64(1), res.OutputArguments[0].(uint64))

		txStatus, err := client.GetTransactionStatus(txId)
		require.NoError(t, err)
		require.EqualValues(t, codec.TRANSACTION_STATUS_COMMITTED, txStatus.TransactionStatus)
	})
}

func Test_SendTransactionAsync(t *testing.T) {
	contractTest(t, func(t *testing.T, endpoint string, vcid uint32) {
		account, _ := orbs.CreateAccount()
		client := orbs.NewClient(endpoint, vcid, codec.NETWORK_TYPE_TEST_NET)

		contractName := deployIncrementContractToGamma(t)

		tx, txId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc", uint32(0))
		require.NoError(t, err)

		res, err := client.SendTransactionAsync(tx)
		require.NoError(t, err)
		require.EqualValues(t, res.RequestStatus, codec.REQUEST_STATUS_IN_PROCESS)
		require.EqualValues(t, res.ExecutionResult, codec.EXECUTION_RESULT_NOT_EXECUTED)
		require.EqualValues(t, res.TransactionStatus, codec.TRANSACTION_STATUS_PENDING)

		time.Sleep(1 * time.Second)
		txStatus, err := client.GetTransactionStatus(txId)
		require.NoError(t, err)
		require.EqualValues(t, codec.TRANSACTION_STATUS_COMMITTED, txStatus.TransactionStatus)
	})
}

func Test_Relay(t *testing.T) {
	contractName := deployIncrementContractToGamma(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fakeTransport := transport.NewMockTransport()
	endpoint := startTrashPanda(ctx, fakeTransport)

	account, _ := orbs.CreateAccount()
	client := orbs.NewClient(endpoint, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	tx, txId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc", uint32(0))
	require.NoError(t, err)

	fakeTransport.Off()

	res, err := client.SendTransaction(tx)
	require.NoError(t, err)
	require.EqualValues(t, res.RequestStatus, codec.REQUEST_STATUS_IN_PROCESS)
	require.EqualValues(t, res.ExecutionResult, codec.EXECUTION_RESULT_NOT_EXECUTED)
	require.EqualValues(t, res.TransactionStatus, codec.TRANSACTION_STATUS_PENDING)

	fakeTransport.On()

	time.Sleep(1 * time.Second)

	txStatus, err := client.GetTransactionStatus(txId)
	require.NoError(t, err)
	require.EqualValues(t, codec.TRANSACTION_STATUS_COMMITTED, txStatus.TransactionStatus)

	notFound, err := client.GetTransactionStatus("0xC0058950d1Bdde15d06C2d7354C3Cb15Dae02CFC6BF5934b358D43dEf1DFE1a0C420Da72e541bd6e")
	require.NoError(t, err)
	require.EqualValues(t, codec.TRANSACTION_STATUS_NO_RECORD_FOUND, notFound.TransactionStatus)
}

func Test_RunQueryWithMultipleEndpoints(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpTransport := transport.NewHttpTransport(transport.Config{
		Timeout: 1 * time.Second,
	})
	endpoint := startTrashPandaWithConfig(ctx, httpTransport, &config.Config{
		Endpoints: []string{
			GAMMA_ENDPOINT,
			BAD_GAMMA_ENDPOINT,
			SECOND_BAD_GAMMA_ENDPOINT,
		},
		VirtualChains:   []uint32{GAMMA_VCHAIN},
		Gamma:           true,
		RelayIntervalMs: 100,
		RelayBatchSize:  10,
		Database:        "./",
	})

	account, _ := orbs.CreateAccount()
	client := orbs.NewClient(endpoint, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	query, err := client.CreateQuery(account.PublicKey, "_Info", "isAlive")
	require.NoError(t, err)

	res, err := client.SendQuery(query)
	require.NoError(t, err)
	require.EqualValues(t, res.RequestStatus, codec.REQUEST_STATUS_COMPLETED)
	require.GreaterOrEqual(t, res.BlockHeight, uint64(1))

}
