package cmd

import (
	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

type BtcConfig struct {
	Network string `mapstructure:"network"`

	CheckpointTag string `mapstructure:"checkpoint-tag"`

	BaseHeader string `mapstructure:"base-header"`

	BaseHeaderHeight uint64 `mapstructure:"base-header-height"`
}

func defaultBabylonBtcConfig() BtcConfig {
	header, height := bbn.GetDefaultBaseHeader()
	return BtcConfig{
		Network:          string(bbn.BtcMainnet),
		CheckpointTag:    string(txformat.MainTag),
		BaseHeader:       header.MarshalHex(),
		BaseHeaderHeight: height,
	}
}

type BabylonAppConfig struct {
	serverconfig.Config `mapstructure:",squash"`

	BtcConfig BtcConfig `mapstructure:"btc-config"`
}

func DefaultBabylonConfig() *BabylonAppConfig {
	return &BabylonAppConfig{
		Config:    *serverconfig.DefaultConfig(),
		BtcConfig: defaultBabylonBtcConfig(),
	}
}

func DefaultBabylonTemplate() string {
	return serverconfig.DefaultConfigTemplate + `
###############################################################################
###                      Babylon Bitcoin configuration                      ###
###############################################################################

[btc-config]

# Configures which bitcoin network should be used for checkpointing
# valid values are: [mainnet, testnet, simnet]
network = "{{ .BtcConfig.Network }}"

# Configures what tag should be prepended to op_return data in btc transaction
# for it to be considered as valid babylon checkpoint
# valid values are:
# "BBT" for testing
# "BBN" for production usage
checkpoint-tag = "{{ .BtcConfig.CheckpointTag }}"

# Configures the base bitcoin header hex that is going to be used
# as the root of the bitcoin chain maintained by babylon
base-header = "{{ .BtcConfig.BaseHeader }}"

# Configures the base bitcoin header height
base-header-height = "{{ .BtcConfig.BaseHeaderHeight }}"
`
}
