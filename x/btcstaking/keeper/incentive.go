package keeper

import (
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) RecordRewardDistCache(ctx sdk.Context) {
	// get BTC tip height and w, which are necessary for determining a BTC
	// delegation's voting power
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		return
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	rdc := types.NewRewardDistCache()

	// iterate all BTC validators to add each BTC validator's distribution info
	// to reward distribution cache
	btcValIter := k.btcValidatorStore(ctx).Iterator(nil, nil)
	defer btcValIter.Close()
	for ; btcValIter.Valid(); btcValIter.Next() {
		valBTCPKBytes := btcValIter.Key()
		valBTCPK, err := bbn.NewBIP340PubKey(valBTCPKBytes)
		if err != nil {
			// failing to unmarshal BTC validator PK in KVStore is a programming error
			panic(err)
		}
		btcVal, err := k.GetBTCValidator(ctx, valBTCPKBytes)
		if err != nil {
			// failing to get a BTC validator with voting power is a programming error
			panic(err)
		}
		if btcVal.IsSlashed() {
			// slashed BTC validator will not get any reward
			continue
		}

		// iterate over all BTC delegations under this validator to compute
		// the BTC validator's distribution info
		btcValDistInfo := types.NewBTCValDistInfo(btcVal)
		btcDelIter := k.btcDelegationStore(ctx, valBTCPK).Iterator(nil, nil)
		for ; btcDelIter.Valid(); btcDelIter.Next() {
			// unmarshal
			var btcDels types.BTCDelegatorDelegations
			k.cdc.MustUnmarshal(btcDelIter.Value(), &btcDels)
			// process each of the BTC delegation
			for _, btcDel := range btcDels.Dels {
				btcValDistInfo.AddBTCDel(btcDel, btcTipHeight, wValue)
			}
		}
		btcDelIter.Close()

		// try to add this BTC validator distribution info to reward distribution cache
		rdc.AddBTCValDistInfo(btcValDistInfo)
	}

	// all good, set the reward distribution cache of the current height
	k.setRewardDistCache(ctx, uint64(ctx.BlockHeight()), rdc)
}

func (k Keeper) setRewardDistCache(ctx sdk.Context, height uint64, rdc *types.RewardDistCache) {
	store := k.rewardDistCacheStore(ctx)
	store.Set(sdk.Uint64ToBigEndian(height), k.cdc.MustMarshal(rdc))
}

func (k Keeper) GetRewardDistCache(ctx sdk.Context, height uint64) (*types.RewardDistCache, error) {
	store := k.rewardDistCacheStore(ctx)
	rdcBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(rdcBytes) == 0 {
		return nil, types.ErrRewardDistCacheNotFound
	}
	var rdc types.RewardDistCache
	k.cdc.MustUnmarshal(rdcBytes, &rdc)
	return &rdc, nil
}

func (k Keeper) RemoveRewardDistCache(ctx sdk.Context, height uint64) {
	store := k.rewardDistCacheStore(ctx)
	store.Delete(sdk.Uint64ToBigEndian(height))
}

// rewardDistCacheStore returns the KVStore of the reward distribution cache
// prefix: RewardDistCacheKey
// key: Babylon block height
// value: RewardDistCache
func (k Keeper) rewardDistCacheStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.RewardDistCacheKey)
}
