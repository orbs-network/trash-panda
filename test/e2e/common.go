package e2e

import (
	"context"
	"github.com/orbs-network/trash-panda/boostrap"
	"os"
	"testing"
)

func GetEndpoint() (endpoint string) {
	endpoint = "http://localhost:8080"

	if endpointFromEnv := os.Getenv("ENDPOINT"); endpointFromEnv != "" {
		endpoint = endpointFromEnv
	}

	return
}

func contractTest(t *testing.T, f func(t *testing.T, endpoint string, vcid uint32)) {
	t.Run("gamma", func(t *testing.T) {
		f(t, "http://localhost:8080", 42)
	})

	t.Run("trash panda", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		boostrap.NewTrashPanda(ctx, 42)
		f(t, "http://localhost:9876/vchains/42", 42)
	})
}
