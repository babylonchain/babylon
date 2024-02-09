package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

// AddBTCDelegation indexes the given BTC delegation in the BTC delegator store, and saves
// it under BTC delegation store
func (k Keeper) AddBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) error {
	if err := btcDel.ValidateBasic(); err != nil {
		return err
	}

	// get staking tx hash
	stakingTxHash, err := btcDel.GetStakingTxHash()
	if err != nil {
		return err
	}

	// for each finality provider the delegation restakes to, update its index
	for _, fpBTCPK := range btcDel.FpBtcPkList {
		var btcDelIndex = types.NewBTCDelegatorDelegationIndex()
		if k.hasBTCDelegatorDelegations(ctx, &fpBTCPK, btcDel.BtcPk) {
			btcDelIndex, err = k.getBTCDelegatorDelegationIndex(ctx, &fpBTCPK, btcDel.BtcPk)
			if err != nil {
				// this can only be a programming error
				panic(fmt.Errorf("failed to get BTC delegations while hasBTCDelegatorDelegations returns true"))
			}
		}

		// index staking tx hash of this BTC delegation
		if err := btcDelIndex.Add(stakingTxHash); err != nil {
			return types.ErrInvalidStakingTx.Wrapf(err.Error())
		}
		// save the index
		store := k.btcDelegatorStore(ctx, &fpBTCPK)
		delBTCPKBytes := btcDel.BtcPk.MustMarshal()
		btcDelIndexBytes := k.cdc.MustMarshal(btcDelIndex)
		store.Set(delBTCPKBytes, btcDelIndexBytes)
	}

	// save this BTC delegation
	k.setBTCDelegation(ctx, btcDel)

	return nil
}

// IterateBTCDelegations iterates all BTC delegations under a given finality provider
func (k Keeper) IterateBTCDelegations(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, handler func(btcDel *types.BTCDelegation) bool) {
	btcDelIter := k.btcDelegatorStore(ctx, fpBTCPK).Iterator(nil, nil)
	defer btcDelIter.Close()
	for ; btcDelIter.Valid(); btcDelIter.Next() {
		// unmarshal delegator's delegation index
		var btcDelIndex types.BTCDelegatorDelegationIndex
		k.cdc.MustUnmarshal(btcDelIter.Value(), &btcDelIndex)
		// retrieve and process each of the BTC delegation
		for _, stakingTxHashBytes := range btcDelIndex.StakingTxHashList {
			stakingTxHash, err := chainhash.NewHash(stakingTxHashBytes)
			if err != nil {
				panic(err) // only programming error is possible
			}
			btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
			shouldContinue := handler(btcDel)
			if !shouldContinue {
				return
			}
		}
	}
}

// hasBTCDelegatorDelegations checks if the given BTC delegator has any BTC delegations under a given finality provider
func (k Keeper) hasBTCDelegatorDelegations(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) bool {
	fpBTCPKBytes := fpBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()

	if !k.HasFinalityProvider(ctx, fpBTCPKBytes) {
		return false
	}
	store := k.btcDelegatorStore(ctx, fpBTCPK)
	return store.Has(delBTCPKBytes)
}

// getBTCDelegatorDelegationIndex gets the BTC delegation index with a given BTC PK under a given finality provider
func (k Keeper) getBTCDelegatorDelegationIndex(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegationIndex, error) {
	fpBTCPKBytes := fpBTCPK.MustMarshal()
	delBTCPKBytes := delBTCPK.MustMarshal()
	store := k.btcDelegatorStore(ctx, fpBTCPK)

	// ensure the finality provider exists
	if !k.HasFinalityProvider(ctx, fpBTCPKBytes) {
		return nil, types.ErrFpNotFound
	}

	// ensure BTC delegator exists
	if !store.Has(delBTCPKBytes) {
		return nil, types.ErrBTCDelegatorNotFound
	}
	// get and unmarshal
	var btcDelIndex types.BTCDelegatorDelegationIndex
	btcDelIndexBytes := store.Get(delBTCPKBytes)
	k.cdc.MustUnmarshal(btcDelIndexBytes, &btcDelIndex)
	return &btcDelIndex, nil
}

// getBTCDelegatorDelegations gets the BTC delegations with a given BTC PK under a given finality provider
func (k Keeper) getBTCDelegatorDelegations(ctx context.Context, fpBTCPK *bbn.BIP340PubKey, delBTCPK *bbn.BIP340PubKey) (*types.BTCDelegatorDelegations, error) {
	btcDelIndex, err := k.getBTCDelegatorDelegationIndex(ctx, fpBTCPK, delBTCPK)
	if err != nil {
		return nil, err
	}
	// get BTC delegation from each staking tx hash
	btcDels := []*types.BTCDelegation{}
	for _, stakingTxHashBytes := range btcDelIndex.StakingTxHashList {
		stakingTxHash, err := chainhash.NewHash(stakingTxHashBytes)
		if err != nil {
			// failing to unmarshal hash bytes in DB's BTC delegation index is a programming error
			panic(err)
		}
		btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
		btcDels = append(btcDels, btcDel)
	}
	return &types.BTCDelegatorDelegations{Dels: btcDels}, nil
}

// GetBTCDelegation gets the BTC delegation with a given staking tx hash
func (k Keeper) GetBTCDelegation(ctx context.Context, stakingTxHashStr string) (*types.BTCDelegation, error) {
	// decode staking tx hash string
	stakingTxHash, err := chainhash.NewHashFromStr(stakingTxHashStr)
	if err != nil {
		return nil, err
	}

	// find BTC delegation from KV store
	btcDel := k.getBTCDelegation(ctx, *stakingTxHash)
	if btcDel == nil {
		return nil, types.ErrBTCDelegationNotFound
	}

	return btcDel, nil
}

// btcDelegatorStore returns the KVStore of the BTC delegators
// prefix: BTCDelegatorKey || finality provider's Bitcoin secp256k1 PK
// key: delegator's Bitcoin secp256k1 PK
// value: BTCDelegatorDelegationIndex (a list of BTCDelegations' staking tx hashes)
func (k Keeper) btcDelegatorStore(ctx context.Context, fpBTCPK *bbn.BIP340PubKey) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	delegationStore := prefix.NewStore(storeAdapter, types.BTCDelegatorKey)
	return prefix.NewStore(delegationStore, fpBTCPK.MustMarshal())
}
