package keeper

import (
	"crypto/sha256"
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

func (k Keeper) setAppHash(ctx sdk.Context, height uint64, appHash []byte) {
	store := k.appHashStore(ctx)
	heightBytes := sdk.Uint64ToBigEndian(height)
	store.Set(heightBytes, appHash)
}

// GetAppHash gets the AppHash of the header at the given height
func (k Keeper) GetAppHash(ctx sdk.Context, height uint64) ([]byte, error) {
	store := k.appHashStore(ctx)
	heightBytes := sdk.Uint64ToBigEndian(height)
	appHash := store.Get(heightBytes)
	if appHash == nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHeight, "height %d is now known in DB yet", height)
	}
	return appHash, nil
}

// RecordAppHash stores the AppHash of the current header to KVStore
func (k Keeper) RecordAppHash(ctx sdk.Context) {
	header := ctx.BlockHeader()
	height := uint64(header.Height)
	k.setAppHash(ctx, height, header.AppHash)
}

// GetAllAppHashsForEpoch fetches all AppHashs in the given epoch
func (k Keeper) GetAllAppHashsForEpoch(ctx sdk.Context, epoch *types.Epoch) ([][]byte, error) {
	// if this epoch is the most recent AND has not ended, then we cannot get all AppHashs for this epoch
	if k.GetEpoch(ctx).EpochNumber == epoch.EpochNumber && !epoch.IsLastBlock(ctx) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHeight, "GetAllAppHashsForEpoch can only be invoked when this epoch has ended")
	}

	// fetch each AppHash in this epoch
	appHashs := [][]byte{}
	for i := epoch.FirstBlockHeight; i <= uint64(epoch.LastBlockHeader.Height); i++ {
		appHash, err := k.GetAppHash(ctx, i)
		if err != nil {
			return nil, err
		}
		appHashs = append(appHashs, appHash)
	}

	return appHashs, nil
}

// ProveAppHashInEpoch generates a proof that the given appHash is in a given epoch
func (k Keeper) ProveAppHashInEpoch(ctx sdk.Context, height uint64, epochNumber uint64) (*tmcrypto.Proof, error) {
	// ensure height is inside this epoch
	epoch, err := k.GetHistoricalEpoch(ctx, epochNumber)
	if err != nil {
		return nil, err
	}
	if !epoch.WithinBoundary(height) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHeight, "the given height %d is not in epoch %d (interval [%d, %d])", height, epoch.EpochNumber, epoch.FirstBlockHeight, uint64(epoch.LastBlockHeader.Height))
	}

	// calculate index of this height in this epoch
	idx := height - epoch.FirstBlockHeight

	// fetch all AppHashs, calculate Merkle tree and proof
	appHashs, err := k.GetAllAppHashsForEpoch(ctx, epoch)
	if err != nil {
		return nil, err
	}
	_, proofs := merkle.ProofsFromByteSlices(appHashs)

	return proofs[idx].ToProto(), nil
}

// VerifyAppHashInclusion verifies whether the given appHash is in the Merkle tree w.r.t. the appHashRoot
func VerifyAppHashInclusion(appHash []byte, appHashRoot []byte, proof *tmcrypto.Proof) error {
	if len(appHash) != sha256.Size {
		return fmt.Errorf("appHash with length %d is not a Sha256 hash", len(appHash))
	}
	if len(appHashRoot) != sha256.Size {
		return fmt.Errorf("appHash with length %d is not a Sha256 hash", len(appHashRoot))
	}
	if proof == nil {
		return fmt.Errorf("proof is nil")
	}

	unwrappedProof, err := merkle.ProofFromProto(proof)
	if err != nil {
		return fmt.Errorf("failed to unwrap proof: %w", err)
	}
	return unwrappedProof.Verify(appHashRoot, appHash)
}

// appHashStore returns the KVStore for the AppHash of each header
// prefix: AppHashKey
// key: height
// value: AppHash in bytes
func (k Keeper) appHashStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.AppHashKey)
}
