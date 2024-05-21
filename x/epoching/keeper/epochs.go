package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/epoching/types"
)

func (k Keeper) setEpochInfo(ctx context.Context, epochNumber uint64, epoch *types.Epoch) {
	store := k.epochInfoStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	epochBytes := k.cdc.MustMarshal(epoch)
	store.Set(epochNumberBytes, epochBytes)
}

func (k Keeper) getEpochInfo(ctx context.Context, epochNumber uint64) (*types.Epoch, error) {
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
func (k Keeper) InitEpoch(ctx context.Context) *types.Epoch {
	header := sdk.UnwrapSDKContext(ctx).HeaderInfo()
	if header.Height > 0 {
		panic("InitEpoch can be invoked only at genesis")
	}
	epochInterval := k.GetParams(ctx).EpochInterval
	epoch := types.NewEpoch(0, epochInterval, 0, &header.Time)
	k.setEpochInfo(ctx, 0, &epoch)
	return &epoch
}

// GetEpoch fetches the current epoch
func (k Keeper) GetEpoch(ctx context.Context) *types.Epoch {
	store := k.epochInfoStore(ctx)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()
	epochBytes := iter.Value()
	var epoch types.Epoch
	k.cdc.MustUnmarshal(epochBytes, &epoch)

	return &epoch
}

func (k Keeper) GetHistoricalEpoch(ctx context.Context, epochNumber uint64) (*types.Epoch, error) {
	epoch, err := k.getEpochInfo(ctx, epochNumber)
	return epoch, err
}

// RecordLastHeaderTime records the last header's timestamp for the current
// epoch, and stores the epoch metadata to KVStore
// The timestamp is used for unbonding delegations once the epoch is timestamped
func (k Keeper) RecordLastHeaderTime(ctx context.Context) error {
	epoch := k.GetEpoch(ctx)
	if !epoch.IsLastBlock(ctx) {
		return errorsmod.Wrapf(types.ErrInvalidHeight, "RecordLastBlockHeader can only be invoked at the last block of an epoch")
	}
	// record last block header
	header := sdk.UnwrapSDKContext(ctx).HeaderInfo()
	epoch.LastBlockTime = &header.Time
	// save back to KVStore
	k.setEpochInfo(ctx, epoch.EpochNumber, epoch)
	return nil
}

// RecordSealerAppHashForPrevEpoch records the AppHash referencing
// the last block of the previous epoch
func (k Keeper) RecordSealerAppHashForPrevEpoch(ctx context.Context) *types.Epoch {
	epoch := k.GetEpoch(ctx)
	if !epoch.IsFirstBlock(ctx) {
		panic(fmt.Errorf("RecordSealerAppHashForPrevEpoch can only be invoked at the first header of a non-zero epoch. "+
			"current epoch: %v, current height: %d", epoch, sdk.UnwrapSDKContext(ctx).HeaderInfo().Height))
	}
	header := sdk.UnwrapSDKContext(ctx).HeaderInfo()

	// get the sealed epoch, i.e., the epoch earlier than the current epoch
	sealedEpoch, err := k.GetHistoricalEpoch(ctx, epoch.EpochNumber-1)
	if err != nil {
		panic(err)
	}

	// record the sealer AppHash for the sealed epoch
	sealedEpoch.SealerAppHash = header.AppHash
	k.setEpochInfo(ctx, sealedEpoch.EpochNumber, sealedEpoch)

	return sealedEpoch
}

// RecordSealerBlockHashForEpoch records the block hash of
// the last block of the current epoch
func (k Keeper) RecordSealerBlockHashForEpoch(ctx context.Context) *types.Epoch {
	// get the sealer header
	epoch := k.GetEpoch(ctx)
	if !epoch.IsLastBlock(ctx) {
		panic(fmt.Errorf("RecordSealerBlockHashForEpoch can only be invoked at the last header of a non-zero epoch. "+
			"current epoch: %v, current height: %d", epoch, sdk.UnwrapSDKContext(ctx).HeaderInfo().Height))
	}
	header := sdk.UnwrapSDKContext(ctx).HeaderInfo()

	// record the sealer block hash for the sealing epoch
	epoch.SealerBlockHash = header.Hash
	k.setEpochInfo(ctx, epoch.EpochNumber, epoch)

	return epoch
}

// IncEpoch adds epoch number by 1
// CONTRACT: can only be invoked at the first block of an epoch
func (k Keeper) IncEpoch(ctx context.Context) types.Epoch {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	epochNumber := k.GetEpoch(ctx).EpochNumber
	incrementedEpochNumber := epochNumber + 1

	epochInterval := k.GetParams(ctx).EpochInterval
	newEpoch := types.NewEpoch(incrementedEpochNumber, epochInterval, uint64(sdkCtx.HeaderInfo().Height), nil)
	k.setEpochInfo(ctx, incrementedEpochNumber, &newEpoch)

	return newEpoch
}

// epochInfoStore returns the store for epoch metadata
// prefix: EpochInfoKey
// key: epochNumber
// value: epoch metadata
func (k Keeper) epochInfoStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.EpochInfoKey)
}
