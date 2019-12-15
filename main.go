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
	"os"
	"time"
)

func main() {
	logger := config.GetLogger()

	configPath := flag.String("config", "./config.json", "path to config")
	showVersion := flag.Bool("version", false, "")
	showConfig := flag.Bool("show-config", false, "")
	flag.Parse()

	if *showVersion {
		fmt.Println(config.GetVersion().String())
		return
	}

	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		logger.Error("config file not found", log.Error(err))
		os.Exit(1)
	}

	cfg, err := config.Parse(configData)
	if err != nil {
		logger.Error("could not parse config file", log.Error(err))
		os.Exit(1)
	}

	if *showConfig {
		fmt.Printf("%+v", *cfg)
		return
	}

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
