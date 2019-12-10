package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/trash-panda/bootstrap"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/transport"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
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

		endpoint := startTrashPanda(ctx, transport.NewHttpTransport())
		f(t, endpoint, GAMMA_VCHAIN)
	})
}

func startTrashPanda(ctx context.Context, transport transport.Transport) string {
	removeDB(GAMMA_VCHAIN)
	httpAddress, endpoint := getRandomAddressAndEnpoint(GAMMA_VCHAIN)
	bootstrap.NewTrashPanda(ctx, transport, &config.Config{
		HttpAddress: httpAddress,
		Endpoints: []string{
			GAMMA_ENDPOINT,
		},
		VirtualChains: []uint32{GAMMA_VCHAIN},
		Gamma:         true,
	})
	return endpoint
}

func getRandomAddressAndEnpoint(vcid uint32) (httpAddress string, endpoint string) {
	rand.Seed(time.Now().UnixNano())
	httpAddress = fmt.Sprintf("localhost:%d", rand.Int31n(63000))
	endpoint = fmt.Sprintf("http://%s/vchains/%d", httpAddress, vcid)

	return
}

func deployIncrementContractToGamma(t *testing.T) (contractName string) {
	account, _ := orbs.CreateAccount()
	return deployIncrementContract(t, account, GAMMA_ENDPOINT, GAMMA_VCHAIN)
}

func deployIncrementContract(t *testing.T, account *orbs.OrbsAccount, endpoint string, vchain uint32) (contractName string) {
	client := orbs.NewClient(endpoint, vchain, codec.NETWORK_TYPE_TEST_NET)

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

func removeDB(vcid uint32) {
	os.RemoveAll(fmt.Sprintf("./vchain-%d.bolt", vcid))
}
