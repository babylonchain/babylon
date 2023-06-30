package keeper

import (
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//nolint:unused
func (k Keeper) setPubRand(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey, height uint64, pr *bbn.SchnorrPubRand) {
	store := k.pubRandStore(ctx, valBtcPK)
	store.Set(sdk.Uint64ToBigEndian(height), *pr)
}

func (k Keeper) HasPubRand(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey, height uint64) bool {
	store := k.pubRandStore(ctx, valBtcPK)
	return store.Has(sdk.Uint64ToBigEndian(height))
}

func (k Keeper) GetPubRand(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey, height uint64) (*bbn.SchnorrPubRand, error) {
	store := k.pubRandStore(ctx, valBtcPK)
	prBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(prBytes) == 0 {
		return nil, types.ErrPubRandNotFound
	}
	return bbn.NewSchnorrPubRand(prBytes)
}

// pubRandStore returns the KVStore of the public randomness
// prefix: PubRandKey
// key: (BTC validator || PK block height)
// value: PublicRandomness
func (k Keeper) pubRandStore(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	prefixedStore := prefix.NewStore(store, types.PubRandKey)
	return prefix.NewStore(prefixedStore, valBtcPK.MustMarshal())
}
