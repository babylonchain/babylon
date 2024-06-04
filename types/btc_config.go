package types

import (
	"math/big"

	"github.com/btcsuite/btcd/chaincfg"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

type SupportedBtcNetwork string

type BtcConfig struct {
	btcNetParams *chaincfg.Params
}

const (
	BtcMainnet SupportedBtcNetwork = "mainnet"
	BtcTestnet SupportedBtcNetwork = "testnet"
	BtcSimnet  SupportedBtcNetwork = "simnet"
	BtcRegtest SupportedBtcNetwork = "regtest"
	BtcSignet  SupportedBtcNetwork = "signet"
)

func getParams(opts servertypes.AppOptions) *chaincfg.Params {
	valueInterface := opts.Get("btc-config.network")

	if valueInterface == nil {
		panic("Bitcoin network should be provided in options")
	}

	network, err := cast.ToStringE(valueInterface)

	if err != nil {
		panic("Bitcoin network config should be valid string")
	}

	if network == string(BtcMainnet) {
		return &chaincfg.MainNetParams
	} else if network == string(BtcTestnet) {
		return &chaincfg.TestNet3Params
	} else if network == string(BtcSimnet) {
		return &chaincfg.SimNetParams
	} else if network == string(BtcRegtest) {
		return &chaincfg.RegressionNetParams
	} else if network == string(BtcSignet) {
		return &chaincfg.SigNetParams
	} else {
		panic("Bitcoin network should be one of [mainet, testnet, simnet, regtest, signet]")
	}
}

func ParseBtcOptionsFromConfig(opts servertypes.AppOptions) BtcConfig {
	return BtcConfig{
		btcNetParams: getParams(opts),
	}
}

func (c *BtcConfig) NetParams() *chaincfg.Params {
	return c.btcNetParams
}

func (c *BtcConfig) PowLimit() big.Int {
	return *c.btcNetParams.PowLimit
}

func (c *BtcConfig) RetargetAdjustmentFactor() int64 {
	return c.btcNetParams.RetargetAdjustmentFactor
}

func (c *BtcConfig) ReduceMinDifficulty() bool {
	return c.btcNetParams.ReduceMinDifficulty
}
