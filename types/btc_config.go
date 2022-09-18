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
	powLimit                 *big.Int
	checkPointTag            txformat.BabylonTag
	retargetAdjustmentFactor int64
	reduceMinDifficulty      bool
}

const (
	BtcMainnet SupportedBtcNetwork = "mainnet"
	BtcTestnet SupportedBtcNetwork = "testnet"
	BtcSimnet  SupportedBtcNetwork = "simnet"
)

func getParams(opts servertypes.AppOptions) chaincfg.Params {
	valueInterface := opts.Get("btc-config.network")

	if valueInterface == nil {
		panic("Bitcoin network should be provided in options")
	}

	network, err := cast.ToStringE(valueInterface)

	if err != nil {
		panic("Bitcoin netowrk config should be valid string")
	}

	if network == string(BtcMainnet) {
		return chaincfg.MainNetParams
	} else if network == string(BtcTestnet) {
		return chaincfg.TestNet3Params
	} else if network == string(BtcSimnet) {
		return chaincfg.SimNetParams
	} else {
		panic("Bitcoin network should be one of [mainet, testnet, simnet]")
	}
}

func parsePowLimit(opts servertypes.AppOptions) *big.Int {
	return getParams(opts).PowLimit
}

func parseRetargetAdjustmentFactor(opts servertypes.AppOptions) int64 {
	return getParams(opts).RetargetAdjustmentFactor
}

func parseReduceMinDifficulty(opts servertypes.AppOptions) bool {
	return getParams(opts).ReduceMinDifficulty
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

	tagBytes := []byte(tag)

	if len(tagBytes) != txformat.TagLength {
		panic("provided tag should have exactly 4 bytes")
	}

	return txformat.BabylonTag(tagBytes)
}

func ParseBtcOptionsFromConfig(opts servertypes.AppOptions) BtcConfig {
	powLimit := parsePowLimit(opts)
	tag := parseCheckpointTag(opts)
	retargetAdjustmentFactor := parseRetargetAdjustmentFactor(opts)
	reduceMinDifficulty := parseReduceMinDifficulty(opts)
	return BtcConfig{
		powLimit:                 powLimit,
		retargetAdjustmentFactor: retargetAdjustmentFactor,
		reduceMinDifficulty:      reduceMinDifficulty,
		checkPointTag:            tag,
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

func (c *BtcConfig) RetargetAdjustmentFactor() int64 {
	return c.retargetAdjustmentFactor
}

func (c *BtcConfig) ReduceMinDifficulty() bool {
	return c.reduceMinDifficulty
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

func GetGlobalRetargetAdjustmentFactor() int64 {
	return btcConfig.RetargetAdjustmentFactor()
}

func GetGlobalReduceMinDifficulty() bool {
	return btcConfig.ReduceMinDifficulty()
}
