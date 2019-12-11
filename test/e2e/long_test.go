package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/trash-panda/config"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

const MAX_BATCHES = 10
const BATCH_SIZE = 50
const INTERVAL = 5 * time.Second

const TESTNET_VCHAIN = 1003

const DEMONET_NODE1 = "http://node1.demonet.orbs.com"
const DEMONET_NODE2 = "http://node2.demonet.orbs.com"
const DEMONET_NODE3 = "http://node3.demonet.orbs.com"
const DEMONET_NODE4 = "http://node4.demonet.orbs.com"

const TEST_CONFIG = "./config.test.json"

func removeConfig() {
	os.RemoveAll(TEST_CONFIG)
}

func skipUnlessTestnet(t *testing.T) {
	if os.Getenv("TESTNET_ENABLED") != "true" {
		t.Skip("testnet disabled")
	}
}

func Test_LongRun(t *testing.T) {
	skipUnlessTestnet(t)

	removeDB(1003)
	removeConfig()

	httpAddress, endpoint := getRandomAddressAndEnpoint(TESTNET_VCHAIN)

	rawJSON, _ := json.Marshal(config.Config{
		HttpAddress:   httpAddress,
		VirtualChains: []uint32{TESTNET_VCHAIN},
		Endpoints: []string{
			DEMONET_NODE1,
			DEMONET_NODE2,
			DEMONET_NODE3,
			DEMONET_NODE4,
		},
		EndpointTimeoutMs: 5000, // 5s
	})
	ioutil.WriteFile(TEST_CONFIG, rawJSON, 0644)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := NewMain(ctx)
	err := m.start()
	require.NoError(t, err)
	defer m.stop()

	time.Sleep(3 * time.Second)

	account, _ := orbs.CreateAccount()

	contractName := deployIncrementContract(t, account, endpoint, TESTNET_VCHAIN)
	print(contractName)

	client := orbs.NewClient(endpoint, TESTNET_VCHAIN, codec.NETWORK_TYPE_TEST_NET)

	txSent := uint64(0)
	for b := 0; b < MAX_BATCHES; b++ {
		var wg sync.WaitGroup

		for i := 0; i < BATCH_SIZE; i++ {
			wg.Add(1)

			go func(nonce uint32) {
				rawTx, _, _ := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "inc", nonce)
				fmt.Println(contractName, "inc", nonce)
				response, err := client.SendTransaction(rawTx)

				var status = ""
				if response != nil {
					status = response.TransactionStatus.String()
					txSent++
				} else if err != nil {
					status = err.Error()
				}

				fmt.Printf("%d/%d [%d/%d] %s\n", i, BATCH_SIZE, b, MAX_BATCHES, status)

				wg.Done()
			}(uint32(b + i))
		}

		wg.Wait()
		<-time.After(INTERVAL)
	}

	require.Eventually(t, func() bool {
		query, _ := client.CreateQuery(account.PublicKey, contractName, "value")
		res, err := client.SendQuery(query)
		if err != nil {
			return false
		}

		println("value: ", res.OutputArguments[0].(uint64), "/", txSent, "/", MAX_BATCHES*BATCH_SIZE)
		return res.OutputArguments[0].(uint64) == txSent
	}, 3*time.Minute, 1*time.Second)

	m.stop()
}

type main struct {
	cmd *exec.Cmd
}

func NewMain(ctx context.Context) *main {
	return &main{
		cmd: exec.CommandContext(ctx, "go", "run", "../../main.go", "--config", TEST_CONFIG),
	}
}

func (m *main) start() error {
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	return m.cmd.Start()
}

func (m *main) stop() error {
	return m.cmd.Process.Kill()
}
