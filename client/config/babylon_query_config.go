package config

import (
	"fmt"
	"net/url"
	"time"
)

// BabylonConfig defines configuration for the Babylon query client
type BabylonQueryConfig struct {
	RPCAddr string        `mapstructure:"rpc-addr"`
	Timeout time.Duration `mapstructure:"timeout"`
}

func (cfg *BabylonQueryConfig) Validate() error {
	if _, err := url.Parse(cfg.RPCAddr); err != nil {
		return fmt.Errorf("cfg.RPCAddr is not correctly formatted: %w", err)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("cfg.Timeout must be positive")
	}
	return nil
}

func DefaultBabylonQueryConfig() BabylonQueryConfig {
	return BabylonQueryConfig{
		RPCAddr: "http://localhost:26657",
		Timeout: 20 * time.Second,
	}
}
