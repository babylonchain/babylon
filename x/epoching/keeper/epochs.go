package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/merkle"
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
	epoch := types.NewEpoch(0, epochInterval, 0, &header)
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

// RecordLastHeaderAndAppHashRoot records the last header and Merkle root of all AppHashs
// for the current epoch, and stores the epoch metadata to KVStore
func (k Keeper) RecordLastHeaderAndAppHashRoot(ctx sdk.Context) error {
	epoch := k.GetEpoch(ctx)
	if !epoch.IsLastBlock(ctx) {
		return errorsmod.Wrapf(types.ErrInvalidHeight, "RecordLastBlockHeader can only be invoked at the last block of an epoch")
	}
	// record last block header
	header := ctx.BlockHeader()
	epoch.LastBlockHeader = &header
	// calculate and record the Merkle root
	appHashs, err := k.GetAllAppHashsForEpoch(ctx, epoch)
	if err != nil {
		return err
	}
	appHashRoot := merkle.HashFromByteSlices(appHashs)
	epoch.AppHashRoot = appHashRoot
	// save back to KVStore
	k.setEpochInfo(ctx, epoch.EpochNumber, epoch)
	return nil
}

// RecordSealerHeaderForPrevEpoch records the sealer header for the previous epoch,
// where the sealer header of an epoch is the 2nd header of the next epoch
// This validator set of the epoch has generated a BLS multisig on `last_commit_hash` of the sealer header
func (k Keeper) RecordSealerHeaderForPrevEpoch(ctx sdk.Context) *types.Epoch {
	// get the sealer header
	epoch := k.GetEpoch(ctx)
	if !epoch.IsSecondBlock(ctx) {
		panic("RecordSealerHeaderForPrevEpoch can only be invoked at the second header of a non-zero epoch")
	}
	header := ctx.BlockHeader()

	// get the sealed epoch, i.e., the epoch earlier than the current epoch
	sealedEpoch, err := k.GetHistoricalEpoch(ctx, epoch.EpochNumber-1)
	if err != nil {
		panic(err)
	}

	// record the sealer header for the sealed epoch
	sealedEpoch.SealerHeader = &header
	k.setEpochInfo(ctx, sealedEpoch.EpochNumber, sealedEpoch)

	return sealedEpoch
}

// IncEpoch adds epoch number by 1
// CONTRACT: can only be invoked at the first block of an epoch
func (k Keeper) IncEpoch(ctx sdk.Context) types.Epoch {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	incrementedEpochNumber := epochNumber + 1
	k.setEpochNumber(ctx, incrementedEpochNumber)

	epochInterval := k.GetParams(ctx).EpochInterval
	newEpoch := types.NewEpoch(incrementedEpochNumber, epochInterval, uint64(ctx.BlockHeight()), nil)
	k.setEpochInfo(ctx, incrementedEpochNumber, &newEpoch)

	return newEpoch
}

// epochInfoStore returns the store for epoch metadata
// prefix: EpochInfoKey
// key: epochNumber
// value: epoch metadata
func (k Keeper) epochInfoStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.EpochInfoKey)
}
