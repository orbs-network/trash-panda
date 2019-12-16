package bootstrap

import (
	"github.com/orbs-network/trash-panda/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_buildProxyConfig(t *testing.T) {
	generalConfig := &config.Config{
		Endpoints: []string{
			"http://node1",
			"http://node2",
			"http://node3",
			"http://node4",
		},
	}
	vcConfig := buildProxyConfig(generalConfig, 1000)

	require.EqualValues(t, []string{
		"http://node1",
		"http://node2",
		"http://node3",
		"http://node4",
	}, generalConfig.Endpoints)

	require.EqualValues(t, []string{
		"http://node1/vchains/1000",
		"http://node2/vchains/1000",
		"http://node3/vchains/1000",
		"http://node4/vchains/1000",
	}, vcConfig.Endpoints)
}
