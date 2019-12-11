package config

import (
	"encoding/json"
)

type Config struct {
	HttpAddress       string
	Gamma             bool
	Endpoints         []string
	EndpointTimeoutMs uint // milliseconds
	VirtualChains     []uint32

	RelayIntervalMs uint
	RelayBatchSize  uint
}

var defaultConfig = Config{
	Endpoints:         []string{"http://localhost:8080"},
	VirtualChains:     []uint32{42},
	HttpAddress:       "localhost:9876",
	EndpointTimeoutMs: 60000,

	RelayIntervalMs: 100,
	RelayBatchSize:  100,
}

func Parse(input []byte) (*Config, error) {
	cfg := &Config{}
	err := json.Unmarshal(input, cfg)

	if len(cfg.VirtualChains) == 0 {
		cfg.VirtualChains = defaultConfig.VirtualChains
	}

	if len(cfg.Endpoints) == 0 {
		cfg.Endpoints = defaultConfig.Endpoints
	}

	if cfg.HttpAddress == "" {
		cfg.HttpAddress = defaultConfig.HttpAddress
	}

	if cfg.EndpointTimeoutMs == 0 {
		cfg.EndpointTimeoutMs = defaultConfig.EndpointTimeoutMs
	}

	if cfg.RelayIntervalMs == 0 {
		cfg.RelayIntervalMs = defaultConfig.RelayIntervalMs
	}

	if cfg.RelayBatchSize == 0 {
		cfg.RelayBatchSize = defaultConfig.RelayBatchSize
	}

	return cfg, err
}
