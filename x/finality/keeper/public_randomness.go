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

// SetPubRandList sets a list of public randomness starting from a given startHeight
// for a given finality provider
func (k Keeper) SetPubRandList(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, startHeight uint64, pubRandList []bbn.SchnorrPubRand) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, writeCache := sdkCtx.CacheContext()

	// write to a KV store cache
	store := k.pubRandFpStore(cacheCtx, fpBtcPK)
	for i, pr := range pubRandList {
		height := startHeight + uint64(i)
		store.Set(sdk.Uint64ToBigEndian(height), pr)
	}

	// atomically write the new public randomness back to KV store
	writeCache()
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

func (k Keeper) IsFirstPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) bool {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)

	// if the iterator is not valid, then this finality provider does not commit any randomness
	return !iter.Valid()
}

// GetLastPubRand retrieves the last public randomness committed by the given finality provider
func (k Keeper) GetLastPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) (uint64, *bbn.SchnorrPubRand, error) {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)

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
// key: (finality provider || PK block height)
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
