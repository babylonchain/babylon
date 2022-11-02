package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func (k Keeper) InsertFork(ctx sdk.Context, chainID string, blocks []*types.IndexedHeader) {
	store := k.forkStore(ctx, chainID)
	fork := types.Fork{
		ChainId: chainID,
		Blocks:  blocks,
	}
	forkBytes := k.cdc.MustMarshal(&fork)
	store.Set(sdk.Uint64ToBigEndian(blocks[0].Height), forkBytes)
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
