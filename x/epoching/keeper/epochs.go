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

// GetEpochNumber fetches the current epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		panic(types.ErrUnknownEpochNumber)
	}
	var epochNumber sdk.Uint
	if err := epochNumber.Unmarshal(bz); err != nil {
		panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
	}

	return epochNumber.Uint64()
}

// GetEpoch fetches the current epoch
func (k Keeper) GetEpoch(ctx sdk.Context) types.Epoch {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		panic(types.ErrUnknownEpochNumber)
	}
	var epochNumber sdk.Uint
	if err := epochNumber.Unmarshal(bz); err != nil {
		panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
	}

	return types.Epoch{
		EpochNumber:   epochNumber.Uint64(),
		EpochInterval: k.GetParams(ctx).EpochInterval,
	}
}

// setEpochNumber sets epoch number
func (k Keeper) setEpochNumber(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes, err := sdk.NewUint(epochNumber).Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}

	store.Set(types.EpochNumberKey, epochNumberBytes)
}

// IncEpochNumber adds epoch number by 1
func (k Keeper) IncEpochNumber(ctx sdk.Context) uint64 {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	incrementedEpochNumber := epochNumber + 1
	k.setEpochNumber(ctx, incrementedEpochNumber)
	return incrementedEpochNumber
}
