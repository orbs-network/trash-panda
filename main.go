package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/bootstrap"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/transport"
	"io/ioutil"
	"time"
)

func main() {
	configPath := flag.String("config", "./config.json", "path to config")
	showVersion := flag.Bool("version", false, "")
	flag.Parse()

	if *showVersion {
		fmt.Println(config.GetVersion().String())
		return
	}

	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(configData)
	}

	cfg, err := config.Parse(configData)
	if err != nil {
		panic(err)
	}

	logger := config.GetLogger()
	logger.Info("starting indexer service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := bootstrap.NewTrashPanda(ctx, transport.NewHttpTransport(transport.Config{
		Timeout: time.Duration(cfg.EndpointTimeoutMs) * time.Millisecond,
	}), cfg)
	if err != nil {
		logger.Error("failed to start the service", log.Error(err))
		cancel()
	} else {
		server.WaitUntilShutdown(ctx)
	}
}
