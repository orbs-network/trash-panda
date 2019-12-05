package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/trash-panda/boostrap"
	"github.com/orbs-network/trash-panda/proxy/adapter/transparent"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

const GAMMA_ENDPOINT = "http://localhost:8080"
const GAMMA_VCHAIN = 42

func contractTest(t *testing.T, f func(t *testing.T, endpoint string, vcid uint32)) {
	t.Run("gamma", func(t *testing.T) {
		f(t, GAMMA_ENDPOINT, GAMMA_VCHAIN)
	})

	t.Run("trash panda", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		httpAddress := fmt.Sprintf("localhost:%d", rand.Intn(63000))
		endpoint := fmt.Sprintf("http://%s/vchains/42", httpAddress)
		boostrap.NewTrashPanda(ctx, transparent.NewTransparentAdapter, httpAddress, 42)
		f(t, endpoint, GAMMA_VCHAIN)
	})
}

func deployIncrementContractToGamma(t *testing.T) (contractName string) {
	account, _ := orbs.CreateAccount()
	client := orbs.NewClient(GAMMA_ENDPOINT, GAMMA_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	sources, err := orbs.ReadSourcesFromDir("../_contracts/counter")
	require.NoError(t, err)

	contractName = fmt.Sprintf("Inc%d", time.Now().UnixNano())
	tx, _, err := client.CreateDeployTransaction(account.PublicKey, account.PrivateKey, contractName, orbs.PROCESSOR_TYPE_NATIVE, sources...)
	require.NoError(t, err)

	res, err := client.SendTransaction(tx)
	require.NoError(t, err)
	require.EqualValues(t, res.ExecutionResult, codec.EXECUTION_RESULT_SUCCESS)

	return
}
