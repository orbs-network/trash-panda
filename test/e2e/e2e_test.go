package e2e

import (
	"context"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
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
		require.GreaterOrEqual(t, res.BlockHeight, uint64(1))
	})

}

func Test_SendTransaction(t *testing.T) {
	contractTest(t, func(t *testing.T, endpoint string, vcid uint32) {
		account, _ := orbs.CreateAccount()
		client := orbs.NewClient(endpoint, vcid, codec.NETWORK_TYPE_TEST_NET)

		contractName := deployIncrementContractToGamma(t)

		tx, txId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc")
		require.NoError(t, err)

		res, err := client.SendTransaction(tx)
		require.NoError(t, err)
		require.EqualValues(t, res.ExecutionResult, codec.EXECUTION_RESULT_SUCCESS)
		require.EqualValues(t, uint64(1), res.OutputArguments[0].(uint64))

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

	tx, txId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc")
	require.NoError(t, err)

	fakeTransport.Off()

	_, err = client.SendTransaction(tx)
	require.Error(t, err)

	fakeTransport.On()

	time.Sleep(1 * time.Second)

	txStatus, err := client.GetTransactionStatus(txId)
	require.NoError(t, err)
	require.EqualValues(t, codec.TRANSACTION_STATUS_COMMITTED, txStatus.TransactionStatus)
}
