package types

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/blockchain"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CalcWork(header *bbl.BTCHeaderBytes) sdk.Uint {
	return sdk.NewUintFromBigInt(blockchain.CalcWork(header.Bits()))
}

func CumulativeWork(childWork sdk.Uint, parentWork sdk.Uint) sdk.Uint {
	sum := sdk.NewUint(0)
	sum = sum.Add(childWork)
	sum = sum.Add(parentWork)
	return sum
}
