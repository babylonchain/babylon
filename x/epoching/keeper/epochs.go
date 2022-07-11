package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultEpochNumber = 0
)

// setEpochNumber sets epoch number
func (k Keeper) setEpochNumber(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Set(types.EpochNumberKey, epochNumberBytes)
}

// InitEpoch sets the zero epoch number to DB
func (k Keeper) InitEpoch(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(0)
	store.Set(types.EpochNumberKey, epochNumberBytes)
}

// GetEpoch fetches the current epoch
func (k Keeper) GetEpoch(ctx sdk.Context) types.Epoch {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		panic(types.ErrUnknownEpochNumber)
	}
	epochNumber := sdk.BigEndianToUint64(bz)
	epochInterval := k.GetParams(ctx).EpochInterval
	return types.Epoch{
		EpochNumber:          epochNumber,
		CurrentEpochInterval: epochInterval,
		FirstBlockHeight:     firstBlockHeight(epochNumber, epochInterval),
	}
}

// IncEpoch adds epoch number by 1
func (k Keeper) IncEpoch(ctx sdk.Context) types.Epoch {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	incrementedEpochNumber := epochNumber + 1
	k.setEpochNumber(ctx, incrementedEpochNumber)
	epochInterval := k.GetParams(ctx).EpochInterval
	return types.Epoch{
		EpochNumber:          incrementedEpochNumber,
		CurrentEpochInterval: epochInterval,
		FirstBlockHeight:     firstBlockHeight(incrementedEpochNumber, epochInterval),
	}
}

// firstBlockHeight returns the height of the first block of a given epoch and epoch interval
// TODO (non-urgent): add support to variable epoch interval
func firstBlockHeight(epochNumber uint64, epochInterval uint64) uint64 {
	// example: in epoch 2, epoch interval is 5 blocks, FirstBlockHeight will be (2-1)*5+1 = 6
	// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
	// 0 |     1     |     2      |
	if epochNumber == 0 {
		return 0
	} else {
		return (epochNumber-1)*epochInterval + 1
	}
}
