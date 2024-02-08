package cmd

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	bbn "github.com/babylonchain/babylon/types"
)

type BtcConfig struct {
	Network string `mapstructure:"network"`
}

func defaultBabylonBtcConfig() BtcConfig {
	return BtcConfig{
		Network: string(bbn.BtcMainnet),
	}
}

type BabylonAppConfig struct {
	serverconfig.Config `mapstructure:",squash"`

	Wasm wasmtypes.WasmConfig `mapstructure:"wasm"`

	BtcConfig BtcConfig `mapstructure:"btc-config"`
}

func DefaultBabylonConfig() *BabylonAppConfig {
	return &BabylonAppConfig{
		Config:    *serverconfig.DefaultConfig(),
		Wasm:      wasmtypes.DefaultWasmConfig(),
		BtcConfig: defaultBabylonBtcConfig(),
	}
}

func DefaultBabylonTemplate() string {
	return serverconfig.DefaultConfigTemplate + wasmtypes.DefaultConfigTemplate() + `
###############################################################################
###                      Babylon Bitcoin configuration                      ###
###############################################################################

[btc-config]

# Configures which bitcoin network should be used for checkpointing
# valid values are: [mainnet, testnet, simnet, signet, regtest]
network = "{{ .BtcConfig.Network }}"
`
}
