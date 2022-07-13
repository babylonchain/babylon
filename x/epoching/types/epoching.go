package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (e Epoch) GetLastBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	return e.FirstBlockHeight + e.CurrentEpochInterval - 1
}

func (e Epoch) GetSecondBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	return e.FirstBlockHeight + 1
}

func (e Epoch) IsLastBlock(ctx sdk.Context) bool {
	return e.GetLastBlockHeight() == uint64(ctx.BlockHeight())
}

func (e Epoch) IsFirstBlock(ctx sdk.Context) bool {
	return e.FirstBlockHeight == uint64(ctx.BlockHeight())
}

func (e Epoch) IsSecondBlock(ctx sdk.Context) bool {
	return e.GetSecondBlockHeight() == uint64(ctx.BlockHeight())
}

func (e Epoch) IsFirstBlockOfNextEpoch(ctx sdk.Context) bool {
	if e.EpochNumber == 0 {
		return ctx.BlockHeight() == 1
	} else {
		height := uint64(ctx.BlockHeight())
		return e.FirstBlockHeight+e.CurrentEpochInterval == height
	}
}
