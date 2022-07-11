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
	return types.Epoch{
		EpochNumber:   epochNumber,
		EpochInterval: k.GetParams(ctx).EpochInterval,
	}
}

// IncEpoch adds epoch number by 1
func (k Keeper) IncEpoch(ctx sdk.Context) types.Epoch {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	incrementedEpochNumber := epochNumber + 1
	k.setEpochNumber(ctx, incrementedEpochNumber)
	return types.Epoch{
		EpochNumber:   incrementedEpochNumber,
		EpochInterval: k.GetParams(ctx).EpochInterval,
	}
}
