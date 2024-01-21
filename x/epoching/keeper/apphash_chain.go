package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cometbft/cometbft/crypto/merkle"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/epoching/types"
)

func (k Keeper) setAppHash(ctx context.Context, height uint64, appHash []byte) {
	store := k.appHashStore(ctx)
	heightBytes := sdk.Uint64ToBigEndian(height)
	store.Set(heightBytes, appHash)
}

// GetAppHash gets the AppHash of the header at the given height
func (k Keeper) GetAppHash(ctx context.Context, height uint64) ([]byte, error) {
	store := k.appHashStore(ctx)
	heightBytes := sdk.Uint64ToBigEndian(height)
	appHash := store.Get(heightBytes)
	if appHash == nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidHeight, "height %d is not known in DB yet", height)
	}
	return appHash, nil
}

// RecordAppHash stores the AppHash of the current header to KVStore
func (k Keeper) RecordAppHash(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := uint64(sdkCtx.HeaderInfo().Height)
	appHash := sdkCtx.HeaderInfo().AppHash
	// HACK: the app hash for the first height is set to nil
	// instead of the hash of an empty byte slice as intended
	// see proposed fix: https://github.com/cosmos/cosmos-sdk/pull/18524
	if height == 1 {
		// $ echo -n '' | sha256sum
		// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		emptyHash := sha256.Sum256([]byte{})
		appHash = emptyHash[:]
	}
	k.setAppHash(ctx, height, appHash)
}

// GetAllAppHashesForEpoch fetches all AppHashes in the given epoch
func (k Keeper) GetAllAppHashesForEpoch(ctx context.Context, epoch *types.Epoch) ([][]byte, error) {
	// if this epoch is the most recent AND has not ended, then we cannot get all AppHashs for this epoch
	if k.GetEpoch(ctx).EpochNumber == epoch.EpochNumber && !epoch.IsLastBlock(sdk.UnwrapSDKContext(ctx)) {
		return nil, errorsmod.Wrapf(types.ErrInvalidHeight, "GetAllAppHashesForEpoch can only be invoked when this epoch has ended")
	}

	// fetch each AppHash in this epoch
	appHashs := [][]byte{}
	for i := epoch.FirstBlockHeight; i <= epoch.GetLastBlockHeight(); i++ {
		appHash, err := k.GetAppHash(ctx, i)
		if err != nil {
			return nil, err
		}
		appHashs = append(appHashs, appHash)
	}

	return appHashs, nil
}

// ProveAppHashInEpoch generates a proof that the given appHash is in a given epoch
func (k Keeper) ProveAppHashInEpoch(ctx context.Context, height uint64, epochNumber uint64) (*tmcrypto.Proof, error) {
	// ensure height is inside this epoch
	epoch, err := k.GetHistoricalEpoch(ctx, epochNumber)
	if err != nil {
		return nil, err
	}
	if !epoch.WithinBoundary(height) {
		return nil, errorsmod.Wrapf(types.ErrInvalidHeight, "the given height %d is not in epoch %d (interval [%d, %d])", height, epoch.EpochNumber, epoch.FirstBlockHeight, epoch.GetLastBlockHeight())
	}

	// calculate index of this height in this epoch
	idx := height - epoch.FirstBlockHeight

	// fetch all AppHashs, calculate Merkle tree and proof
	appHashs, err := k.GetAllAppHashesForEpoch(ctx, epoch)
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
func (k Keeper) appHashStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.AppHashKey)
}
