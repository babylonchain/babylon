package types

import (
	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/btcsuite/btcd/chaincfg"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
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
	BtcRegtest SupportedBtcNetwork = "regtest"
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
	} else if network == string(BtcRegtest) {
		return chaincfg.RegressionNetParams
	} else {
		panic("Bitcoin network should be one of [mainet, testnet, simnet, regtest]")
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
