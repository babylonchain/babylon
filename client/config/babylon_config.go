package config

import (
	"fmt"
	"github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// BabylonConfig defines configuration for the Babylon client
// adapted from https://github.com/strangelove-ventures/lens/blob/v0.5.1/client/config.go
type BabylonConfig struct {
	Key              string        `mapstructure:"key"`
	ChainID          string        `mapstructure:"chain-id"`
	RPCAddr          string        `mapstructure:"rpc-addr"`
	GRPCAddr         string        `mapstructure:"grpc-addr"`
	AccountPrefix    string        `mapstructure:"account-prefix"`
	KeyringBackend   string        `mapstructure:"keyring-backend"`
	GasAdjustment    float64       `mapstructure:"gas-adjustment"`
	GasPrices        string        `mapstructure:"gas-prices"`
	KeyDirectory     string        `mapstructure:"key-directory"`
	Debug            bool          `mapstructure:"debug"`
	Timeout          time.Duration `mapstructure:"timeout"`
	BlockTimeout     time.Duration `mapstructure:"block-timeout"`
	OutputFormat     string        `mapstructure:"output-format"`
	SignModeStr      string        `mapstructure:"sign-mode"`
	SubmitterAddress string        `mapstructure:"submitter-address"`
}

func (cfg *BabylonConfig) Validate() error {
	if _, err := url.Parse(cfg.RPCAddr); err != nil {
		return fmt.Errorf("rpc-addr is not correctly formatted: %w", err)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if cfg.BlockTimeout < 0 {
		return fmt.Errorf("block-timeout can't be negative")
	}
	return nil
}

func (cfg *BabylonConfig) ToCosmosProviderConfig() cosmos.CosmosProviderConfig {
	return cosmos.CosmosProviderConfig{
		Key:            cfg.Key,
		ChainID:        cfg.ChainID,
		RPCAddr:        cfg.RPCAddr,
		AccountPrefix:  cfg.AccountPrefix,
		KeyringBackend: cfg.KeyringBackend,
		GasAdjustment:  cfg.GasAdjustment,
		GasPrices:      cfg.GasPrices,
		KeyDirectory:   cfg.KeyDirectory,
		Debug:          cfg.Debug,
		Timeout:        cfg.Timeout.String(),
		BlockTimeout:   cfg.BlockTimeout.String(),
		OutputFormat:   cfg.OutputFormat,
		SignModeStr:    cfg.SignModeStr,
	}
}

func DefaultBabylonConfig() BabylonConfig {
	return BabylonConfig{
		Key:     "node0",
		ChainID: "chain-test",
		// see https://docs.cosmos.network/master/core/grpc_rest.html for default ports
		// TODO: configure HTTPS for Babylon's RPC server
		// TODO: how to use Cosmos SDK's RPC server (port 1317) rather than Tendermint's RPC server (port 26657)?
		RPCAddr: "http://localhost:26657",
		// TODO: how to support GRPC in the Babylon client?
		GRPCAddr:         "https://localhost:9090",
		AccountPrefix:    "bbn",
		KeyringBackend:   "test",
		GasAdjustment:    1.2,
		GasPrices:        "0.01ubbn",
		KeyDirectory:     defaultBabylonHome(),
		Debug:            true,
		Timeout:          20 * time.Second,
		OutputFormat:     "json",
		SignModeStr:      "direct",
		SubmitterAddress: "bbn1v6k7k9s8md3k29cu9runasstq5zaa0lpznk27w", // this is currently a placeholder, will not recognized by Babylon
	}
}

// defaultBabylonHome returns the default Babylon node directory, which is $HOME/.babylond
// copied from https://github.com/babylonchain/babylon/blob/648b804bc492ded2cb826ba261d7164b4614d78a/app/app.go#L205-L210
func defaultBabylonHome() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(userHomeDir, ".babylond")
}
