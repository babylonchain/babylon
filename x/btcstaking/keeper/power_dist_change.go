package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/* power distribution update */

// UpdatePowerDist updates the voting power distribution of finality providers
// and their BTC delegations
func (k Keeper) UpdatePowerDist(ctx context.Context) {
	params := k.GetParams(ctx)
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// prepare metrics for {active, inactive} finality providers,
	// {pending, active, unbonded BTC delegations}, and total staked Bitcoins
	// NOTE: slashed finality providers and BTC delegations are recorded upon
	// slashing events rather than here
	var (
		numFPs        int    = 0
		numStakedSats uint64 = 0
		numDelsMap           = map[types.BTCDelegationStatus]int{
			types.BTCDelegationStatus_PENDING:  0,
			types.BTCDelegationStatus_ACTIVE:   0,
			types.BTCDelegationStatus_UNBONDED: 0,
		}
	)
	// for voting power and rewards
	dc := types.NewVotingPowerDistCache()

	// iterate over all finality providers to find out non-slashed ones that have
	// positive voting power
	k.IterateActiveFPs(
		ctx,
		func(fp *types.FinalityProvider) bool {
			fpDistInfo := types.NewFinalityProviderDistInfo(fp)

			// iterate over all BTC delegations under the finality provider
			// in order to accumulate voting power dist info for it
			k.IterateBTCDelegations(ctx, fp.BtcPk, func(btcDel *types.BTCDelegation) bool {
				// accumulate voting power and reward distribution cache
				fpDistInfo.AddBTCDel(btcDel, btcTipHeight, wValue, params.CovenantQuorum)

				// record metrics
				numStakedSats += btcDel.VotingPower(btcTipHeight, wValue, params.CovenantQuorum)
				numDelsMap[btcDel.GetStatus(btcTipHeight, wValue, params.CovenantQuorum)]++

				return true
			})

			if fpDistInfo.TotalVotingPower > 0 {
				dc.AddFinalityProviderDistInfo(fpDistInfo)
			}

			return true
		},
	)
	// record metrics for finality providers and total staked BTCs
	numActiveFPs := min(numFPs, int(params.MaxActiveFinalityProviders))
	types.RecordActiveFinalityProviders(numActiveFPs)
	types.RecordInactiveFinalityProviders(numFPs - numActiveFPs)
	numStakedBTCs := float32(numStakedSats / SatoshisPerBTC)
	types.RecordMetricsKeyStakedBitcoins(numStakedBTCs)
	// record metrics for BTC delegations
	for status, num := range numDelsMap {
		types.RecordBTCDelegations(num, status)
	}

	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	maxNumActiveFPs := k.GetParams(ctx).MaxActiveFinalityProviders
	activeFps := types.FilterTopNFinalityProviders(dc.FinalityProviders, maxNumActiveFPs)
	// set voting power table and re-calculate total voting power of top N finality providers
	dc.TotalVotingPower = uint64(0)
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	for _, fp := range activeFps {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.TotalVotingPower)
		dc.TotalVotingPower += fp.TotalVotingPower
	}

	// set the voting power distribution cache of the current height
	k.setVotingPowerDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), dc)
}

/* finality provider slash events storage */

// setFinalityProviderEvent sets adds a given slashed finality provider's
// BTC PK to finality provider events storage
func (k Keeper) setFinalityProviderEvent(ctx context.Context, fpBTCPK []byte) {
	store := k.finalityProviderEventStore(ctx)
	// NOTE: value is currently never used so doesn't matter
	store.Set(fpBTCPK, []byte("slashed"))
}

// removeFinalityProviderEvents removes all finality provider events
// This is called after processing all finality provider events in `BeginBlocker`
func (k Keeper) removeFinalityProviderEvents(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	storeAdapter.Delete(types.FinalityProviderEventKey)
}

// iterateFinalityProviderEvents uses the given handler function to handle
// all finality provider events
// This is called in `BeginBlocker`
func (k Keeper) iterateFinalityProviderEvents(
	ctx context.Context,
	handleFunc func(fpBTCPK []byte) bool,
) {
	store := k.finalityProviderEventStore(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		fpBTCPK := iter.Key()
		shouldContinue := handleFunc(fpBTCPK)
		if !shouldContinue {
			break
		}
	}
}

// finalityProviderEventStore returns the KVStore of the finality provider events
// key: FinalityProviderEventKey
// value: finality provider's BTC PK
// value: event (current it can only be slashed)
func (k Keeper) finalityProviderEventStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.FinalityProviderEventKey)
}

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
