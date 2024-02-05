package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) RecordRewardDistCache(ctx context.Context) {
	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	// get BTC tip height and w, which are necessary for determining a BTC
	// delegation's voting power
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		return
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	fpDistMap := map[string]*types.FinalityProviderDistInfo{}

	k.IterateActiveFPsAndBTCDelegations(
		ctx,
		func(fp *types.FinalityProvider, btcDel *types.BTCDelegation) {
			fpBTCPKHex := fp.BtcPk.MarshalHex()
			// create fp dist info if not exist
			if _, ok := fpDistMap[fpBTCPKHex]; !ok {
				fpDistMap[fpBTCPKHex] = types.NewFinalityProviderDistInfo(fp)
			}
			// append BTC delegation
			fpDistMap[fpBTCPKHex].AddBTCDel(btcDel, btcTipHeight, wValue, covenantQuorum)
		},
	)

	// create reward distribution cache
	rdc := types.NewRewardDistCache()
	for _, fpDistInfo := range fpDistMap {
		// try to add this finality provider distribution info to reward distribution cache
		rdc.AddFinalityProviderDistInfo(fpDistInfo)
	}

	// all good, set the reward distribution cache of the current height
	k.setRewardDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), rdc)
}

func (k Keeper) setRewardDistCache(ctx context.Context, height uint64, rdc *types.RewardDistCache) {
	store := k.rewardDistCacheStore(ctx)
	store.Set(sdk.Uint64ToBigEndian(height), k.cdc.MustMarshal(rdc))
}

func (k Keeper) GetRewardDistCache(ctx context.Context, height uint64) (*types.RewardDistCache, error) {
	store := k.rewardDistCacheStore(ctx)
	rdcBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(rdcBytes) == 0 {
		return nil, types.ErrRewardDistCacheNotFound
	}
	var rdc types.RewardDistCache
	k.cdc.MustUnmarshal(rdcBytes, &rdc)
	return &rdc, nil
}

func (k Keeper) RemoveRewardDistCache(ctx context.Context, height uint64) {
	store := k.rewardDistCacheStore(ctx)
	store.Delete(sdk.Uint64ToBigEndian(height))
}

// rewardDistCacheStore returns the KVStore of the reward distribution cache
// prefix: RewardDistCacheKey
// key: Babylon block height
// value: RewardDistCache
func (k Keeper) rewardDistCacheStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.RewardDistCacheKey)
}
