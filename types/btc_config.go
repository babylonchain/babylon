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
	powLimit         *big.Int
	checkPointTag    txformat.BabylonTag
	baseHeader       BTCHeaderBytes
	baseHeaderHeight uint64
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

func parseBaseHex(opts servertypes.AppOptions) BTCHeaderBytes {
	valueInterface := opts.Get("btc-config.base-header")
	if valueInterface == nil {
		panic("Bitcoin base header should be provided in options")
	}

	headerHex, err := cast.ToStringE(valueInterface)

	if err != nil {
		panic("base-header should be a valid string")
	}

	headerBytes, err := NewBTCHeaderBytesFromHex(headerHex)

	if err != nil {
		panic("base-header should be a valid header hex")
	}
	return headerBytes
}

func parseBaseHeaderHeight(opts servertypes.AppOptions) uint64 {
	valueInterface := opts.Get("btc-config.base-header-height")
	if valueInterface == nil {
		panic("Bitcoin base header height should be provided in options")
	}

	headerHeight, err := cast.ToUint64E(valueInterface)

	if err != nil {
		panic("base-header-height should be a valid uint64")
	}

	return headerHeight
}

func ParseBtcOptionsFromConfig(opts servertypes.AppOptions) BtcConfig {
	powLimit := parsePowLimit(opts)
	tag := parseCheckpointTag(opts)
	baseHex := parseBaseHex(opts)
	baseHeight := parseBaseHeaderHeight(opts)
	return BtcConfig{
		powLimit:         powLimit,
		checkPointTag:    tag,
		baseHeader:       baseHex,
		baseHeaderHeight: baseHeight,
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

func (c *BtcConfig) BaseHeader() BTCHeaderBytes {
	return c.baseHeader
}

func (c *BtcConfig) BaseHeaderHeight() uint64 {
	return c.baseHeaderHeight
}

func GetDefaultBaseHeader() (BTCHeaderBytes, uint64) {
	hex := "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22"
	headerBytes, err := NewBTCHeaderBytesFromHex(hex)
	if err != nil {
		panic("Invalid default base header hex")
	}
	return headerBytes, 736056
}

func GetBaseBTCHeaderHeight() uint64 {
	return btcConfig.BaseHeaderHeight()
}

func GetBaseBTCHeader() BTCHeaderBytes {
	return btcConfig.BaseHeader()
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
