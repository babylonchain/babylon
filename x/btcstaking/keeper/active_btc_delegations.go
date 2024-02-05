package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// indexActiveBTCDelegation adds an active BTC delegation to the BTC delegator delegation index
func (k Keeper) indexActiveBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) {
	stakingTxHash := btcDel.MustGetStakingTxHash()

	// for each finality provider the delegation restakes to, update its index
	for _, fpBTCPK := range btcDel.FpBtcPkList {
		// get active BTC delegation index from KV store
		var btcDelIndex = k.getActiveBTCDelegationIndex(ctx, &fpBTCPK, btcDel.BtcPk)
		if btcDelIndex == nil {
			btcDelIndex = types.NewBTCDelegatorDelegationIndex()
		}
		// index staking tx hash of this BTC delegation
		if err := btcDelIndex.Add(stakingTxHash); err != nil {
			panic(types.ErrInvalidStakingTx.Wrapf(err.Error()))
		}
		// save the index back to KV store
		k.setActiveBTCDelegationIndex(ctx, &fpBTCPK, btcDel.BtcPk, btcDelIndex)
	}
}

// setActiveBTCDelegationIndex set the active BTC delegation index with a given BTC PK under a given finality provider to KV store
func (k Keeper) setActiveBTCDelegationIndex(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey, btcDelIndex *types.BTCDelegatorDelegationIndex) {
	store := k.activeBTCDelegationStore(ctx, fpBTCPK)
	delBTCPKBytes := delBTCPK.MustMarshal()
	btcDelIndexBytes := k.cdc.MustMarshal(btcDelIndex)
	store.Set(delBTCPKBytes, btcDelIndexBytes)
}

// getActiveBTCDelegationIndex gets the active BTC delegation index with a given BTC PK under a given finality provider
func (k Keeper) getActiveBTCDelegationIndex(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) *types.BTCDelegatorDelegationIndex {
	delBTCPKBytes := delBTCPK.MustMarshal()
	store := k.activeBTCDelegationStore(ctx, fpBTCPK)

	// get and unmarshal
	var btcDelIndex types.BTCDelegatorDelegationIndex
	btcDelIndexBytes := store.Get(delBTCPKBytes)
	if len(btcDelIndexBytes) == 0 {
		return nil
	}
	k.cdc.MustUnmarshal(btcDelIndexBytes, &btcDelIndex)
	return &btcDelIndex
}

// btcDelegatorStore returns the KVStore of the BTC delegators
// prefix: ActiveBTCDelegationKey || finality provider's Bitcoin secp256k1 PK
// key: delegator's Bitcoin secp256k1 PK
// value: BTCDelegatorDelegationIndex (a list of BTCDelegations' staking tx hashes)
func (k Keeper) activeBTCDelegationStore(ctx context.Context, fpBTCPK *bbn.BIP340PubKey) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	delegationStore := prefix.NewStore(storeAdapter, types.BTCDelegatorKey)
	return prefix.NewStore(delegationStore, fpBTCPK.MustMarshal())
}
