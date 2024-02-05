package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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

	rdc := types.NewRewardDistCache()

	// iterate all finality providers to add each finality provider's distribution info
	// to reward distribution cache
	fpIter := k.finalityProviderStore(ctx).Iterator(nil, nil)
	defer fpIter.Close()
	for ; fpIter.Valid(); fpIter.Next() {
		fpBTCPKBytes := fpIter.Key()
		fpBTCPK, err := bbn.NewBIP340PubKey(fpBTCPKBytes)
		if err != nil {
			// failing to unmarshal finality provider PK in KVStore is a programming error
			panic(err)
		}
		fp, err := k.GetFinalityProvider(ctx, fpBTCPKBytes)
		if err != nil {
			// failing to get a finality provider with voting power is a programming error
			panic(err)
		}
		if fp.IsSlashed() {
			// slashed finality provider will not get any reward
			continue
		}

		fpDistInfo := types.NewFinalityProviderDistInfo(fp)

		// iterate over all BTC delegations under this finality provider to compute
		// the finality provider's distribution info
		// wrap it inside a function to prevent corrupt DB state
		// https://stackoverflow.com/questions/45617758/proper-way-to-release-resources-with-defer-in-a-loop/45620423
		func() {
			btcDelIter := k.btcDelegatorStore(ctx, fpBTCPK).Iterator(nil, nil)
			defer btcDelIter.Close()
			for ; btcDelIter.Valid(); btcDelIter.Next() {
				// unmarshal
				var btcDelIndex types.BTCDelegatorDelegationIndex
				k.cdc.MustUnmarshal(btcDelIter.Value(), &btcDelIndex)
				// retrieve and process each of the BTC delegation
				for _, stakingTxHashBytes := range btcDelIndex.StakingTxHashList {
					stakingTxHash, err := chainhash.NewHash(stakingTxHashBytes)
					if err != nil {
						panic(err) // only programming error is possible
					}
					btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
					fpDistInfo.AddBTCDel(btcDel, btcTipHeight, wValue, covenantQuorum)
				}
			}
		}()

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
