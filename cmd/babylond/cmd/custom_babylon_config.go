package cmd

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"
)

const (
	defaultKeyName       = ""
	defaultGasPrice      = "0.01ubbn"
	defaultGasAdjustment = 1.5
)

type BtcConfig struct {
	Network string `mapstructure:"network"`

	CheckpointTag string `mapstructure:"checkpoint-tag"`
}

func defaultBabylonBtcConfig() BtcConfig {
	return BtcConfig{
		Network:       string(bbn.BtcMainnet),
		CheckpointTag: txformat.DefaultMainTagStr,
	}
}

func defaultSignerConfig() SignerConfig {
	return SignerConfig{
		KeyName:       defaultKeyName,
		GasPrice:      defaultGasPrice,
		GasAdjustment: defaultGasAdjustment,
	}
}

type SignerConfig struct {
	KeyName       string  `mapstructure:"key-name"`
	GasPrice      string  `mapstructure:"gas-price"`
	GasAdjustment float64 `mapstructure:"gas-adjustment"`
}

type BabylonAppConfig struct {
	serverconfig.Config `mapstructure:",squash"`

	Wasm wasmtypes.WasmConfig `mapstructure:"wasm"`

	BtcConfig BtcConfig `mapstructure:"btc-config"`

	SignerConfig SignerConfig `mapstructure:"signer-config"`
}

func DefaultBabylonConfig() *BabylonAppConfig {
	return &BabylonAppConfig{
		Config:       *serverconfig.DefaultConfig(),
		Wasm:         wasmtypes.DefaultWasmConfig(),
		BtcConfig:    defaultBabylonBtcConfig(),
		SignerConfig: defaultSignerConfig(),
	}
}

func DefaultBabylonTemplate() string {
	return serverconfig.DefaultConfigTemplate + wasmtypes.DefaultConfigTemplate() + `
###############################################################################
###                      Babylon Bitcoin configuration                      ###
###############################################################################

[btc-config]

# Configures which bitcoin network should be used for checkpointing
# valid values are: [mainnet, testnet, simnet, regtest]
network = "{{ .BtcConfig.Network }}"


# Configures what tag should be prepended to op_return data in btc transaction
# for it to be considered as valid babylon checkpoint. Must have exactly 4 bytes.
checkpoint-tag = "{{ .BtcConfig.CheckpointTag }}"

[signer-config]

# Configures which key that the BLS signer uses to sign BLS-sig transactions
key-name = "{{ .SignerConfig.KeyName }}"
# Configures the gas-price that the signer would like to pay
gas-price = "{{ .SignerConfig.GasPrice }}"
# Configures the adjustment of the gas cost of estimation
gas-adjustment = "{{ .SignerConfig.GasAdjustment }}"
`
}
