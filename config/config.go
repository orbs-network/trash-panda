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

	Database string
}

var defaultConfig = Config{
	Endpoints:         []string{"http://localhost:8080"},
	VirtualChains:     []uint32{42},
	HttpAddress:       "localhost:9876",
	EndpointTimeoutMs: 60000,

	RelayIntervalMs: 1000,
	RelayBatchSize:  100,

	Database: "./db/",
}

func Parse(input []byte) (*Config, error) {
	return mergeWithDefaults(input)
}

func mergeWithDefaults(input []byte) (*Config, error) {
	cfg := make(map[string]interface{})
	if err := json.Unmarshal(input, cfg); err != nil {
		return nil, err
	}

	rawDefaults, _ := json.Marshal(defaultConfig)
	defaultCfg := make(map[string]interface{})
	json.Unmarshal(rawDefaults, defaultCfg)

	for k, v := range defaultCfg {
		if _, found := cfg[k]; !found {
			cfg[k] = v
		}
	}

	finalConfigRaw, _ := json.Marshal(cfg)
	finalConfig := &Config{}
	json.Unmarshal(finalConfigRaw, finalConfig)

	return finalConfig, nil
}
