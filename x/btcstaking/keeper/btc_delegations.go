package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/* BTC delegation state update event store */

// setBTCDelegationEvent records the BTC delegation with the given staking tx
// hash enters or will enter the given new state at the given BTC height
func (k Keeper) setBTCDelegationEvent(
	ctx context.Context,
	btcHeight uint64,
	stakingTxHash *chainhash.Hash,
	newState types.BTCDelegationStatus,
) {
	store := k.btcDelegationEventStore(ctx, btcHeight)
	store.Set(stakingTxHash[:], newState.ToBytes())
}

// removeBTCDelegationEvents removes all BTC delegation state update events
// at a given BTC height
// This is called after processing all BTC delegation events in `BeginBlocker`
func (k Keeper) removeBTCDelegationEvents(ctx context.Context, btcHeight uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.BTCDelegationEventKey)
	store.Delete(sdk.Uint64ToBigEndian(btcHeight))
}

// iterateBTCDelegationEvents uses the given handler function to handle each
// BTC delegation state update that happens at the given BTC height
// This is called in `BeginBlocker`
func (k Keeper) iterateBTCDelegationEvents(
	ctx context.Context,
	btcHeight uint64,
	handleFunc func(stakingTxHash *chainhash.Hash, newState *types.BTCDelegationStatus) bool,
) {
	store := k.btcDelegationEventStore(ctx, btcHeight)
	btcDelEventIter := store.Iterator(nil, nil)
	defer btcDelEventIter.Close()
	for ; btcDelEventIter.Valid(); btcDelEventIter.Next() {
		stakingTxHash, err := chainhash.NewHash(btcDelEventIter.Key())
		if err != nil {
			panic(err) // only programming error
		}
		newState, err := types.NewBTCDelegationStatus(btcDelEventIter.Value())
		if err != nil {
			panic(err) // only programming error
		}
		shouldContinue := handleFunc(stakingTxHash, &newState)
		if !shouldContinue {
			break
		}
	}
}

// btcDelegationEventStore returns the KVStore of the state update
// events of BTC delegations
// prefix: BTCDelegationEventKey
// key: (BTC height || BTC delegation's staking tx hash)
// value: BTCDelegationStatus
func (k Keeper) btcDelegationEventStore(ctx context.Context, btcHeight uint64) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.BTCDelegationEventKey)
	return prefix.NewStore(store, sdk.Uint64ToBigEndian(btcHeight))
}

/* BTC delegation store */

func (k Keeper) setBTCDelegation(ctx context.Context, btcDel *types.BTCDelegation) {
	store := k.btcDelegationStore(ctx)
	stakingTxHash := btcDel.MustGetStakingTxHash()
	btcDelBytes := k.cdc.MustMarshal(btcDel)
	store.Set(stakingTxHash[:], btcDelBytes)
}

func (k Keeper) getBTCDelegation(ctx context.Context, stakingTxHash chainhash.Hash) *types.BTCDelegation {
	store := k.btcDelegationStore(ctx)
	btcDelBytes := store.Get(stakingTxHash[:])
	if len(btcDelBytes) == 0 {
		return nil
	}
	var btcDel types.BTCDelegation
	k.cdc.MustUnmarshal(btcDelBytes, &btcDel)
	return &btcDel
}

// btcDelegationStore returns the KVStore of the BTC delegations
// prefix: BTCDelegationKey
// key: BTC delegation's staking tx hash
// value: BTCDelegation
func (k Keeper) btcDelegationStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.BTCDelegationKey)
}
