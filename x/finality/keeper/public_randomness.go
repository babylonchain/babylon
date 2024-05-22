package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
	Public randomness commitment storage
*/

func (k Keeper) IsFirstPubRandCommit(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) bool {
	store := k.pubRandCommitFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	// if the iterator is not valid, then this finality provider does not commit any randomness
	return !iter.Valid()
}

// GetPubRandCommitForHeight finds the public randomness commitment that includes the given
// height for the given finality provider
func (k Keeper) GetPubRandCommitForHeight(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, height uint64) (*types.PubRandCommit, error) {
	store := k.pubRandCommitFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	var prCommit types.PubRandCommit
	for ; iter.Valid(); iter.Next() {
		k.cdc.MustUnmarshal(iter.Value(), &prCommit)
		if prCommit.IsInRange(height) {
			return &prCommit, nil
		}
	}
	return nil, types.ErrPubRandNotFound
}

// SetPubRandCommit adds the given public randomness commitment for the given public key
func (k Keeper) SetPubRandCommit(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, prCommit *types.PubRandCommit) {
	store := k.pubRandCommitFpStore(ctx, fpBtcPK)
	prcBytes := k.cdc.MustMarshal(prCommit)
	store.Set(sdk.Uint64ToBigEndian(prCommit.StartHeight), prcBytes)
}

// GetLastPubRandCommit retrieves the last public randomness commitment of the given finality provider
func (k Keeper) GetLastPubRandCommit(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) *types.PubRandCommit {
	store := k.pubRandCommitFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	if !iter.Valid() {
		// this finality provider does not commit any randomness
		return nil
	}

	var prCommit types.PubRandCommit
	k.cdc.MustUnmarshal(iter.Value(), &prCommit)
	return &prCommit
}

// pubRandCommitFpStore returns the KVStore of the commitment of public randomness
// prefix: PubRandKey
// key: (finality provider PK || block height of the commitment)
// value: PubRandCommit
func (k Keeper) pubRandCommitFpStore(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) prefix.Store {
	store := k.pubRandCommitStore(ctx)
	return prefix.NewStore(store, fpBtcPK.MustMarshal())
}

// pubRandCommitStore returns the KVStore of the public randomness commitments
// prefix: PubRandKey
// key: (prefix)
// value: PubRandCommit
func (k Keeper) pubRandCommitStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.PubRandCommitKey)
}

/*
	Public randomness storage
*/

// SetPubRand sets a public randomness at a given height for a given finality provider
func (k Keeper) SetPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, height uint64, pubRand bbn.SchnorrPubRand) {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	store.Set(sdk.Uint64ToBigEndian(height), pubRand)
}

func (k Keeper) HasPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, height uint64) bool {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	return store.Has(sdk.Uint64ToBigEndian(height))
}

func (k Keeper) GetPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, height uint64) (*bbn.SchnorrPubRand, error) {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	prBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(prBytes) == 0 {
		return nil, types.ErrPubRandNotFound
	}
	return bbn.NewSchnorrPubRand(prBytes)
}

// GetLastPubRand retrieves the last public randomness committed by the given finality provider
func (k Keeper) GetLastPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) (uint64, *bbn.SchnorrPubRand, error) {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	if !iter.Valid() {
		// this finality provider does not commit any randomness
		return 0, nil, types.ErrNoPubRandYet
	}

	height := sdk.BigEndianToUint64(iter.Key())
	pubRand, err := bbn.NewSchnorrPubRand(iter.Value())
	if err != nil {
		// failing to marshal public randomness in KVStore can only be a programming error
		panic(fmt.Errorf("failed to unmarshal public randomness in KVStore: %w", err))
	}
	return height, pubRand, nil
}

// pubRandFpStore returns the KVStore of the public randomness
// prefix: PubRandKey
// key: (finality provider PK || block height)
// value: PublicRandomness
func (k Keeper) pubRandFpStore(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) prefix.Store {
	prefixedStore := k.pubRandStore(ctx)
	return prefix.NewStore(prefixedStore, fpBtcPK.MustMarshal())
}

// pubRandStore returns the KVStore of the public randomness
// prefix: PubRandKey
// key: (prefix)
// value: PublicRandomness
func (k Keeper) pubRandStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.PubRandKey)
}
