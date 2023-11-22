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

func (k Keeper) setPubRand(ctx context.Context, valBtcPK *bbn.BIP340PubKey, height uint64, pr *bbn.SchnorrPubRand) {
	store := k.pubRandStore(ctx, valBtcPK)
	store.Set(sdk.Uint64ToBigEndian(height), *pr)
}

// SetPubRandList sets a list of public randomness starting from a given startHeight
// for a given BTC validator
func (k Keeper) SetPubRandList(ctx context.Context, valBtcPK *bbn.BIP340PubKey, startHeight uint64, pubRandList []bbn.SchnorrPubRand) {
	for i, pr := range pubRandList {
		k.setPubRand(ctx, valBtcPK, startHeight+uint64(i), &pr)
	}
}

func (k Keeper) HasPubRand(ctx context.Context, valBtcPK *bbn.BIP340PubKey, height uint64) bool {
	store := k.pubRandStore(ctx, valBtcPK)
	return store.Has(sdk.Uint64ToBigEndian(height))
}

func (k Keeper) GetPubRand(ctx context.Context, valBtcPK *bbn.BIP340PubKey, height uint64) (*bbn.SchnorrPubRand, error) {
	store := k.pubRandStore(ctx, valBtcPK)
	prBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(prBytes) == 0 {
		return nil, types.ErrPubRandNotFound
	}
	return bbn.NewSchnorrPubRand(prBytes)
}

func (k Keeper) IsFirstPubRand(ctx context.Context, valBtcPK *bbn.BIP340PubKey) bool {
	store := k.pubRandStore(ctx, valBtcPK)
	iter := store.ReverseIterator(nil, nil)

	// if the iterator is not valid, then this validator does not commit any randomness
	return !iter.Valid()
}

// GetLastPubRand retrieves the last public randomness committed by the given BTC validator
func (k Keeper) GetLastPubRand(ctx context.Context, valBtcPK *bbn.BIP340PubKey) (uint64, *bbn.SchnorrPubRand, error) {
	store := k.pubRandStore(ctx, valBtcPK)
	iter := store.ReverseIterator(nil, nil)

	if !iter.Valid() {
		// this validator does not commit any randomness
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

// pubRandStore returns the KVStore of the public randomness
// prefix: PubRandKey
// key: (BTC validator || PK block height)
// value: PublicRandomness
func (k Keeper) pubRandStore(ctx context.Context, valBtcPK *bbn.BIP340PubKey) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixedStore := prefix.NewStore(storeAdapter, types.PubRandKey)
	return prefix.NewStore(prefixedStore, valBtcPK.MustMarshal())
}
