package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetHeader(ctx sdk.Context, chainID string, height uint64) (*types.IndexedHeader, error) {
	store := k.canonicalChainStore(ctx, chainID)
	heightBytes := sdk.Uint64ToBigEndian(height)
	if !store.Has(heightBytes) {
		return nil, types.ErrHeaderNotFound
	}
	headerBytes := store.Get(heightBytes)
	var header types.IndexedHeader
	k.cdc.MustUnmarshal(headerBytes, &header)
	return &header, nil
}

func (k Keeper) InsertHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) error {
	if header == nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header is nil")
	}
	// NOTE: we can accept header without ancestor since IBC connection can be established at any height
	store := k.canonicalChainStore(ctx, chainID)
	store.Set(sdk.Uint64ToBigEndian(header.Height), k.cdc.MustMarshal(header))
	return nil
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
