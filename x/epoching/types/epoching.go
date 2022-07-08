package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (e Epoch) LastBlockHeight() uint64 {
	// example: in epoch 1, epoch interval is 5 blocks, LastBlockHeight will be 1*5=5
	// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
	// 0 |     1     |     2      |
	return e.EpochNumber * e.EpochInterval
}

func (e Epoch) FirstBlockHeight() uint64 {
	if e.EpochNumber == 0 {
		return 0
	}
	// example: in epoch 1, epoch interval is 5 blocks, FirstBlockHeight will be LastBlockHeight-5+1=1
	// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
	// 0 |     1     |     2      |
	return e.LastBlockHeight() - e.EpochInterval + 1
}

func (e Epoch) IsLastBlock(ctx sdk.Context) bool {
	return e.LastBlockHeight() == uint64(ctx.BlockHeight())
}

func (e Epoch) IsFirstBlock(ctx sdk.Context) bool {
	return e.FirstBlockHeight() == uint64(ctx.BlockHeight())
}
