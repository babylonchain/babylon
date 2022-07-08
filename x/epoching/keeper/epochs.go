package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultEpochNumber = 0
)

// setEpochNumber sets epoch number
func (k Keeper) InitEpochNumber(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes, err := sdk.NewUint(0).Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}

	store.Set(types.EpochNumberKey, epochNumberBytes)
}

// GetEpochNumber fetches epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) sdk.Uint {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		panic(types.ErrUnknownEpochNumber)
	}
	var epochNumber sdk.Uint
	if err := epochNumber.Unmarshal(bz); err != nil {
		panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
	}

	return epochNumber
}

// setEpochNumber sets epoch number
func (k Keeper) setEpochNumber(ctx sdk.Context, epochNumber sdk.Uint) {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}

	store.Set(types.EpochNumberKey, epochNumberBytes)
}

// IncEpochNumber adds epoch number by 1
func (k Keeper) IncEpochNumber(ctx sdk.Context) sdk.Uint {
	epochNumber := k.GetEpochNumber(ctx)
	incrementedEpochNumber := epochNumber.AddUint64(1)
	k.setEpochNumber(ctx, incrementedEpochNumber)
	return incrementedEpochNumber
}

// GetEpochBoundary gets the epoch boundary, i.e., the height of the block that ends this epoch
// example: in epoch 1, epoch interval is 5 blocks, boundary will be 1*5=5
// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
// 0 |     1     |     2      |
func (k Keeper) GetEpochBoundary(ctx sdk.Context) sdk.Uint {
	epochNumber := k.GetEpochNumber(ctx)
	// epoch number is 0 at the 0-th block, i.e., genesis
	if epochNumber.IsZero() {
		return sdk.NewUint(0)
	}
	// case when epoch number > 0
	epochInterval := sdk.NewUint(k.GetParams(ctx).EpochInterval)
	return epochNumber.Mul(epochInterval)
}
