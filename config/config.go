package config

import (
	"encoding/json"
)

type Config struct {
	HttpAddress   string
	Gamma         bool
	Endpoints     []string
	VirtualChains []uint32
}

var defaultConfig = Config{
	Endpoints:     []string{"http://localhost:8080"},
	VirtualChains: []uint32{42},
	HttpAddress:   "localhost:9876",
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

	return cfg, err
}
