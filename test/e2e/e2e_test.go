package e2e

import (
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/stretchr/testify/require"
	"testing"
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

		query, _, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc")
		require.NoError(t, err)

		res, err := client.SendTransaction(query)
		require.NoError(t, err)
		require.EqualValues(t, res.ExecutionResult, codec.EXECUTION_RESULT_SUCCESS)
		require.EqualValues(t, uint64(1), res.OutputArguments[0].(uint64))
	})
}
