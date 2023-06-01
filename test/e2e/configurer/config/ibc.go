package config

import (
	"fmt"

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

// NewIBCChannelConfigWithBabylonContract creates an IBC channel between a Babylon chain and
// a Babylon contract deployed on another Cosmos zone
// Note that Babylon contract initiates the IBC channel, thus is the chainA in the config
func NewIBCChannelConfigWithBabylonContract(BabylonChainID string, BabylonContractChainID string, contractAddr string) *IBCChannelConfig {
	return &IBCChannelConfig{
		ChainAID:   BabylonContractChainID,
		ChainBID:   BabylonChainID,
		ChainAPort: fmt.Sprintf("wasm.%s", contractAddr),
		ChainBPort: zctypes.PortID,
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
