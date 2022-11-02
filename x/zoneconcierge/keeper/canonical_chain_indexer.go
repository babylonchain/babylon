package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetHeader(ctx sdk.Context, chainID string, height uint64) *types.IndexedHeader {
	store := k.canonicalChainStore(ctx, chainID)
	headerBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(headerBytes) == 0 {
		return nil
	}
	var header types.IndexedHeader
	k.cdc.MustUnmarshal(headerBytes, &header)
	return &header
}

func (k Keeper) InsertHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) {
	store := k.canonicalChainStore(ctx, chainID)
	store.Set(sdk.Uint64ToBigEndian(header.Height), k.cdc.MustMarshal(header))
}

// canonicalChainStore stores the canonical chain of a CZ, formed as a list of IndexedHeader
// prefix: CanonicalChainKey || chainID
// key: height
// value: IndexedHeader
func (k Keeper) canonicalChainStore(ctx sdk.Context, chainID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	canonicalChainStore := prefix.NewStore(store, types.CanonicalChainKey)
	chainIDBytes := []byte(chainID)
	return prefix.NewStore(canonicalChainStore, chainIDBytes)
}
