package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) setVotingPowerDistCache(ctx context.Context, height uint64, dc *types.VotingPowerDistCache) {
	store := k.votingPowerDistCacheStore(ctx)
	store.Set(sdk.Uint64ToBigEndian(height), k.cdc.MustMarshal(dc))
}

func (k Keeper) GetVotingPowerDistCache(ctx context.Context, height uint64) (*types.VotingPowerDistCache, error) {
	store := k.votingPowerDistCacheStore(ctx)
	rdcBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(rdcBytes) == 0 {
		return nil, types.ErrVotingPowerDistCacheNotFound
	}
	var dc types.VotingPowerDistCache
	k.cdc.MustUnmarshal(rdcBytes, &dc)
	return &dc, nil
}

func (k Keeper) RemoveVotingPowerDistCache(ctx context.Context, height uint64) {
	store := k.votingPowerDistCacheStore(ctx)
	store.Delete(sdk.Uint64ToBigEndian(height))
}

// votingPowerDistCacheStore returns the KVStore of the voting power distribution cache
// prefix: VotingPowerDistCacheKey
// key: Babylon block height
// value: VotingPowerDistCache
func (k Keeper) votingPowerDistCacheStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.VotingPowerDistCacheKey)
}
