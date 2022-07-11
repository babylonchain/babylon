package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (e Epoch) LastBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	return e.FirstBlockHeight + e.CurrentEpochInterval - 1
}

func (e Epoch) SecondBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	return e.FirstBlockHeight + 1
}

func (e Epoch) IsLastBlock(ctx sdk.Context) bool {
	return e.LastBlockHeight() == uint64(ctx.BlockHeight())
}

func (e Epoch) IsFirstBlock(ctx sdk.Context) bool {
	return e.FirstBlockHeight == uint64(ctx.BlockHeight())
}

func (e Epoch) IsSecondBlock(ctx sdk.Context) bool {
	return e.SecondBlockHeight() == uint64(ctx.BlockHeight())
}
