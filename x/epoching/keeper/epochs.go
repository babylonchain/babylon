package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultEpochNumber = 0
)

// setEpoch sets epoch number
func (k Keeper) setEpochNumber(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Set(types.EpochNumberKey, epochNumberBytes)
}

func (k Keeper) getEpochNumber(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		panic(types.ErrUnknownEpochNumber)
	}
	epochNumber := sdk.BigEndianToUint64(bz)
	return epochNumber
}

func (k Keeper) setEpochInfo(ctx sdk.Context, epochNumber uint64, epoch *types.Epoch) {
	store := k.epochInfoStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	epochBytes := k.cdc.MustMarshal(epoch)
	store.Set(epochNumberBytes, epochBytes)
}

func (k Keeper) getEpochInfo(ctx sdk.Context, epochNumber uint64) (*types.Epoch, error) {
	store := k.epochInfoStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	bz := store.Get(epochNumberBytes)
	if bz == nil {
		return nil, types.ErrUnknownEpochNumber
	}
	var epoch types.Epoch
	k.cdc.MustUnmarshal(bz, &epoch)
	return &epoch, nil
}

// InitEpoch sets the zero epoch number to DB
func (k Keeper) InitEpoch(ctx sdk.Context) {
	header := ctx.BlockHeader()
	if header.Height > 0 {
		panic("InitEpoch can be invoked only at genesis")
	}
	epochInterval := k.GetParams(ctx).EpochInterval
	epoch := types.NewEpoch(0, epochInterval, &header)
	k.setEpochInfo(ctx, 0, &epoch)

	k.setEpochNumber(ctx, 0)
}

// GetEpoch fetches the current epoch
func (k Keeper) GetEpoch(ctx sdk.Context) *types.Epoch {
	epochNumber := k.getEpochNumber(ctx)
	epoch, err := k.getEpochInfo(ctx, epochNumber)
	if err != nil {
		panic(err)
	}
	return epoch
}

func (k Keeper) GetHistoricalEpoch(ctx sdk.Context, epochNumber uint64) (*types.Epoch, error) {
	epoch, err := k.getEpochInfo(ctx, epochNumber)
	return epoch, err
}

func (k Keeper) FinalizeEpoch(ctx sdk.Context) *types.Epoch {
	epoch := k.GetEpoch(ctx)
	if !epoch.IsLastBlock(ctx) {
		panic("FinalizeEpoch can only be invoked at the last block of an epoch")
	}
	header := ctx.BlockHeader()
	epoch.LastBlockHeader = &header
	k.setEpochInfo(ctx, epoch.EpochNumber, epoch)
	return epoch
}

// IncEpoch adds epoch number by 1
func (k Keeper) IncEpoch(ctx sdk.Context) types.Epoch {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	incrementedEpochNumber := epochNumber + 1
	k.setEpochNumber(ctx, incrementedEpochNumber)

	epochInterval := k.GetParams(ctx).EpochInterval
	newEpoch := types.NewEpoch(incrementedEpochNumber, epochInterval, nil)
	k.setEpochInfo(ctx, incrementedEpochNumber, &newEpoch)

	return newEpoch
}

func (k Keeper) epochInfoStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.EpochInfoKey)
}
