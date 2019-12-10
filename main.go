package main

import (
	"context"
	"flag"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/boostrap"
	"github.com/orbs-network/trash-panda/bootstrap"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/transport"
	"io/ioutil"
)

func main() {
	logger := config.GetLogger()
	logger.Info("starting indexer service")

	configPath := flag.String("config", "./config.json", "path to config")
	flag.Parse()

	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(configData)
	}

	cfg, err := config.Parse(configData)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := bootstrap.NewTrashPanda(ctx, transport.NewHttpTransport(), cfg)
	if err != nil {
		logger.Error("failed to start the service", log.Error(err))
		cancel()
	} else {
		server.WaitUntilShutdown(ctx)
	}
}
