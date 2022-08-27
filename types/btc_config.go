package types

import (
	"math/big"
	"sync"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/btcsuite/btcd/chaincfg"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

// Global bitcoin configuration
// It is global, as we are validating few things in stateless checkTx, which
// are dependent on the btc network we are using
var (
	btcConfig  *BtcConfig
	initConfig sync.Once
)

type SupportedBtcNetwork string

type BtcConfig struct {
	powLimit      *big.Int
	checkPointTag txformat.BabylonTag
}

const (
	BtcMainnet SupportedBtcNetwork = "mainnet"
	BtcTestnet SupportedBtcNetwork = "testnet"
	BtcSimnet  SupportedBtcNetwork = "simnet"
)

func parsePowLimit(opts servertypes.AppOptions) *big.Int {
	valueInterface := opts.Get("btc-config.network")

	if valueInterface == nil {
		panic("Bitcoin network should be provided in options")
	}

	network, err := cast.ToStringE(valueInterface)

	if err != nil {
		panic("Btcoin netowrk config should be valid string")
	}

	if network == string(BtcMainnet) {
		return chaincfg.MainNetParams.PowLimit
	} else if network == string(BtcTestnet) {
		return chaincfg.TestNet3Params.PowLimit
	} else if network == string(BtcSimnet) {
		return chaincfg.SimNetParams.PowLimit
	} else {
		panic("Bicoin network should be one of [mainet, testnet, simnet]")
	}
}

func parseCheckpointTag(opts servertypes.AppOptions) txformat.BabylonTag {
	valueInterface := opts.Get("btc-config.checkpoint-tag")

	if valueInterface == nil {
		panic("Bitcoin network should be provided in options")
	}

	tag, err := cast.ToStringE(valueInterface)

	if err != nil {
		panic("checkpoint-tag should be valid string")
	}

	if tag == string(txformat.MainTag) {
		return txformat.MainTag
	} else if tag == string(txformat.TestTag) {
		return txformat.TestTag
	} else {
		panic("tag should be one of [BBN, BBT]")
	}

}

func ParseBtcOptionsFromConfig(opts servertypes.AppOptions) BtcConfig {
	powLimit := parsePowLimit(opts)
	tag := parseCheckpointTag(opts)
	return BtcConfig{
		powLimit:      powLimit,
		checkPointTag: tag,
	}
}

func InitGlobalBtcConfig(c BtcConfig) {
	initConfig.Do(func() {
		btcConfig = &c
	})
}

func (c *BtcConfig) PowLimit() big.Int {
	return *c.powLimit
}

func (c *BtcConfig) CheckpointTag() txformat.BabylonTag {
	return c.checkPointTag
}

func GetGlobalPowLimit() big.Int {
	// We are making copy of pow limit to avoid anyone changing globally configured
	// powlimit. If it start slowing things down, due to multiple copies needed to
	// be garbage collected, we will need to think of other way of protecting global
	// state
	return btcConfig.PowLimit()
}

func GetGlobalCheckPointTag() txformat.BabylonTag {
	return btcConfig.checkPointTag
}
