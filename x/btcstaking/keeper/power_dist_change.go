package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
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

	// record voting power distribution and cache
	k.recordVotingPowerAndCache(ctx, dc)
}

func (k Keeper) recordVotingPowerAndCache(ctx context.Context, dc *types.VotingPowerDistCache) {
	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	params := k.GetParams(ctx)
	activeFps := types.FilterTopNFinalityProviders(dc.FinalityProviders, params.MaxActiveFinalityProviders)
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

func (k Keeper) UpdatePowerDist2(ctx context.Context) {
	height := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}

	// at the end, clear all events processed in this function
	defer k.ClearPowerDistUpdateEvents(ctx, btcTipHeight)

	// get the power dist cache
	dc := k.GetVotingPowerDistCache(ctx, height)
	// get all power distirbution update events
	events := k.GetAllPowerDistUpdateEvents(ctx, btcTipHeight)

	if dc == nil && len(events) == 0 {
		// no existing BTC delegation and no event, no nothing
		return
	} else if dc != nil && len(events) == 0 {
		// map everything in prev height to this height
		k.recordVotingPowerAndCache(ctx, dc)
	} else {
		if dc == nil {
			dc = types.NewVotingPowerDistCache()
		}
		// TODO: reconcile voting power distribution cache and new events

		// record voting power and cache for this height
		k.recordVotingPowerAndCache(ctx, dc)
	}
}

/* voting power distribution update event store */

// addPowerDistUpdateEvent appends an event that affect voting power distribution
// to the store
func (k Keeper) addPowerDistUpdateEvent(
	ctx context.Context,
	btcHeight uint64,
	event *types.EventPowerDistUpdate,
) {
	store := k.powerDistUpdateEventStore(ctx, btcHeight)

	// get event index
	eventIdx := uint64(0) // event index starts from 0
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()
	if iter.Valid() {
		// if there exists events already, event index will be the subsequent one
		eventIdx = sdk.BigEndianToUint64(iter.Key()) + 1
	}

	// key is event index, and value is the event bytes
	store.Set(sdk.Uint64ToBigEndian(eventIdx), k.cdc.MustMarshal(event))
}

// ClearPowerDistUpdateEvents removes all BTC delegation state update events
// at a given BTC height
// This is called after processing all BTC delegation events in `BeginBlocker`
// nolint:unused
func (k Keeper) ClearPowerDistUpdateEvents(ctx context.Context, btcHeight uint64) {
	store := k.powerDistUpdateEventStore(ctx, btcHeight)
	keys := [][]byte{}

	// get all keys
	// using an enclosure to ensure iterator is closed right after
	// the function is done
	func() {
		iter := store.Iterator(nil, nil)
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			keys = append(keys, iter.Key())
		}
	}()

	// remove all keys
	for _, key := range keys {
		store.Delete(key)
	}
}

// GetAllPowerDistUpdateEvents gets all voting power update events
func (k Keeper) GetAllPowerDistUpdateEvents(ctx context.Context, btcHeight uint64) []*types.EventPowerDistUpdate {
	events := []*types.EventPowerDistUpdate{}
	k.IteratePowerDistUpdateEvents(ctx, btcHeight, func(event *types.EventPowerDistUpdate) bool {
		events = append(events, event)
		return true
	})
	return events
}

// IteratePowerDistUpdateEvents uses the given handler function to handle each
// voting power distribution update event that happens at the given BTC height.
// This is called in `BeginBlocker`
func (k Keeper) IteratePowerDistUpdateEvents(
	ctx context.Context,
	btcHeight uint64,
	handleFunc func(event *types.EventPowerDistUpdate) bool,
) {
	store := k.powerDistUpdateEventStore(ctx, btcHeight)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var event types.EventPowerDistUpdate
		k.cdc.MustUnmarshal(iter.Value(), &event)
		shouldContinue := handleFunc(&event)
		if !shouldContinue {
			break
		}
	}
}

// powerDistUpdateEventStore returns the KVStore of events that affect
// voting power distribution
// prefix: PowerDistUpdateKey
// key: (BTC height || event index)
// value: BTCDelegationStatus
func (k Keeper) powerDistUpdateEventStore(ctx context.Context, btcHeight uint64) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.PowerDistUpdateKey)
	return prefix.NewStore(store, sdk.Uint64ToBigEndian(btcHeight))
}
