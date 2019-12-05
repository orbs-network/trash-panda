package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/trash-panda/boostrap"
	"github.com/orbs-network/trash-panda/proxy/adapter/transparent"
	"math/rand"
	"testing"
)

func contractTest(t *testing.T, f func(t *testing.T, endpoint string, vcid uint32)) {
	t.Run("gamma", func(t *testing.T) {
		f(t, "http://localhost:8080", 42)
	})

	t.Run("trash panda", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		httpAddress := fmt.Sprintf("localhost:%d", rand.Intn(63000))
		endpoint := fmt.Sprintf("http://%s/vchains/42", httpAddress)
		boostrap.NewTrashPanda(ctx, transparent.NewTransparentAdapter, httpAddress, 42)
		f(t, endpoint, 42)
	})
}
