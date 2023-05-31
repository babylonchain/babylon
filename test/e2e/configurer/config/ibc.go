package config

import (
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

type IBCChannelConfig struct {
	ChainAID   string
	ChainBID   string
	ChainAPort string
	ChainBPort string
	Ordering   channeltypes.Order
	Version    string
}

func NewIBCChannelConfigTwoBabylonChains(chainAID string, chainBID string) *IBCChannelConfig {
	return &IBCChannelConfig{
		ChainAID:   chainAID,
		ChainBID:   chainBID,
		ChainAPort: zctypes.PortID,
		ChainBPort: zctypes.PortID,
		Ordering:   zctypes.Ordering,
		Version:    zctypes.Version,
	}
}

func NewIBCChannelConfigWithBabylonContract(chainAID string, chainBID string, contractAddr string) *IBCChannelConfig {
	return &IBCChannelConfig{
		ChainAID:   chainAID,
		ChainBID:   chainBID,
		ChainAPort: zctypes.PortID,
		ChainBPort: "wasm." + contractAddr,
		Ordering:   zctypes.Ordering,
		Version:    zctypes.Version,
	}
}

func (c *IBCChannelConfig) ToCmd() []string {
	return []string{"hermes", "create", "channel",
		"--a-chain", c.ChainAID, "--b-chain", c.ChainBID, // chain ID
		"--a-port", c.ChainAPort, "--b-port", c.ChainBPort, // port
		"--order", c.Ordering.String(), // ordering
		"--channel-version", c.Version, // version
		"--new-client-connection", "--yes",
	}
}
