package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetFork returns a list of forked headers at a given height
func (k Keeper) GetFork(ctx sdk.Context, chainID string, height uint64) *types.Fork {
	store := k.forkStore(ctx, chainID)
	forkBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(forkBytes) == 0 {
		return nil
	}
	var fork types.Fork
	k.cdc.MustUnmarshal(forkBytes, &fork)
	return &fork
}

// InsertForkHeader inserts a forked header to the list of forked headers at the same height
func (k Keeper) InsertForkHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) {
	store := k.forkStore(ctx, chainID)
	fork := k.GetFork(ctx, chainID, header.Height)
	fork.Headers = append(fork.Headers, header)
	forkBytes := k.cdc.MustMarshal(fork)
	store.Set(sdk.Uint64ToBigEndian(header.Height), forkBytes)
}

// forkStore stores the forks for each CZ
// prefix: ForkKey || chainID
// key: height that this fork starts from
// value: a list of IndexedHeader, representing each header in the fork
func (k Keeper) forkStore(ctx sdk.Context, chainID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	forkStore := prefix.NewStore(store, types.ForkKey)
	chainIDBytes := []byte(chainID)
	return prefix.NewStore(forkStore, chainIDBytes)
}
