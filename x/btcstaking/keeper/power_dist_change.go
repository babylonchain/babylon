package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/* power distribution update */

// UpdatePowerDist updates the voting power table and distribution cache.
// This is triggered upon each `BeginBlock`
func (k Keeper) UpdatePowerDist(ctx context.Context) {
	height := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	btcTipHeight := k.GetCurrentBTCHeight(ctx)

	// get the power dist cache in the last height
	dc := k.GetVotingPowerDistCache(ctx, height-1)
	// get all power distribution update events during the previous tip
	// and the current tip
	lastBTCTipHeight := k.GetBTCHeightAtBabylonHeight(ctx, height-1)
	events := k.GetAllPowerDistUpdateEvents(ctx, lastBTCTipHeight, btcTipHeight)

	// if no event exists, then map previous voting power and
	// cache to the current height
	if len(events) == 0 {
		if dc != nil {
			// map everything in prev height to this height
			k.recordVotingPowerAndCache(ctx, dc)
		}
		return
	}

	// clear all events that have been consumed in this function
	defer func() {
		for i := lastBTCTipHeight; i <= btcTipHeight; i++ {
			k.ClearPowerDistUpdateEvents(ctx, i)
		}
	}()

	if dc == nil {
		// no BTC staker at the prior height
		dc = types.NewVotingPowerDistCache()
	}

	// reconcile old voting power distribution cache and new events
	// to construct the new distribution
	newDc := k.processAllPowerDistUpdateEvents(ctx, dc, events)

	// record voting power and cache for this height
	k.recordVotingPowerAndCache(ctx, newDc)

	// record metrics
	k.recordMetrics(ctx, newDc)
}

func (k Keeper) recordVotingPowerAndCache(ctx context.Context, dc *types.VotingPowerDistCache) {
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)

	// set voting power table for this height
	for _, fp := range dc.ActiveFinalityProviders {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.TotalVotingPower)
	}

	// set the voting power distribution cache of the current height
	k.setVotingPowerDistCache(ctx, babylonTipHeight, dc)
}

func (k Keeper) recordMetrics(ctx context.Context, dc *types.VotingPowerDistCache) {
	// number of active FPs
	numActiveFPs := len(dc.ActiveFinalityProviders)
	types.RecordActiveFinalityProviders(numActiveFPs)
	// number of inactive FPs
	numInactiveFPs := len(dc.FinalityProviders) - numActiveFPs
	types.RecordInactiveFinalityProviders(numInactiveFPs)
	// staked Satoshi
	stakedSats := uint64(0)
	for _, fp := range dc.FinalityProviders {
		stakedSats += fp.TotalVotingPower
	}
	numStakedBTCs := float32(stakedSats / SatoshisPerBTC)
	types.RecordMetricsKeyStakedBitcoins(numStakedBTCs)
	// TODO: record number of BTC delegations under different status
}

// processAllPowerDistUpdateEvents processes all events that affect
// voting power distribution and returns a new distribution cache.
// The following events will affect the voting power distribution:
// - newly active BTC delegations
// - newly unbonded BTC delegations
// - slashed finality providers
func (k Keeper) processAllPowerDistUpdateEvents(
	ctx context.Context,
	dc *types.VotingPowerDistCache,
	events []*types.EventPowerDistUpdate,
) *types.VotingPowerDistCache {
	// a map where key is finality provider's BTC PK hex and value is a list
	// of BTC delegations that newly become active under this provider
	activeBTCDels := map[string][]*types.BTCDelegation{}
	// a map where key is unbonded BTC delegation's staking tx hash
	unbondedBTCDels := map[string]struct{}{}
	// a map where key is slashed finality providers' BTC PK
	slashedFPs := map[string]struct{}{}

	/*
		filter and classify all events into new/expired BTC delegations and slashed FPs
	*/
	for _, event := range events {
		switch typedEvent := event.Ev.(type) {
		case *types.EventPowerDistUpdate_BtcDelStateUpdate:
			delEvent := typedEvent.BtcDelStateUpdate
			if delEvent.NewState == types.BTCDelegationStatus_ACTIVE {
				// newly active BTC delegation
				btcDel, err := k.GetBTCDelegation(ctx, delEvent.StakingTxHash)
				if err != nil {
					panic(err) // only programming error
				}
				// add the BTC delegation to each restaked finality provider
				for _, fpBTCPK := range btcDel.FpBtcPkList {
					fpBTCPKHex := fpBTCPK.MarshalHex()
					activeBTCDels[fpBTCPKHex] = append(activeBTCDels[fpBTCPKHex], btcDel)
				}
			} else if delEvent.NewState == types.BTCDelegationStatus_UNBONDED {
				// add the expired BTC delegation to the map
				unbondedBTCDels[delEvent.StakingTxHash] = struct{}{}
			}
		case *types.EventPowerDistUpdate_SlashedFp:
			// slashed finality providers
			slashedFPs[typedEvent.SlashedFp.Pk.MarshalHex()] = struct{}{}
		}
	}

	// if no voting power update, return directly
	noUpdate := len(activeBTCDels) == 0 && len(unbondedBTCDels) == 0 && len(slashedFPs) == 0
	if noUpdate {
		return dc
	}

	/*
		At this point, there is voting power update.
		Then, construct a voting power dist cache by reconciling the previous
		cache and all the new events.
	*/
	newDc := types.NewVotingPowerDistCache()

	// iterate over all finality providers and apply all events
	for i := range dc.FinalityProviders {
		// create a copy of the finality provider
		fp := *dc.FinalityProviders[i]
		fp.TotalVotingPower = 0
		fp.BtcDels = []*types.BTCDelDistInfo{}

		fpBTCPKHex := fp.BtcPk.MarshalHex()

		// if this finality provider is slashed, continue to avoid recording it
		if _, ok := slashedFPs[fpBTCPKHex]; ok {
			continue
		}

		// add all BTC delegations that are not unbonded to the new finality provider
		for j := range dc.FinalityProviders[i].BtcDels {
			btcDel := *dc.FinalityProviders[i].BtcDels[j]
			if _, ok := unbondedBTCDels[btcDel.StakingTxHash]; !ok {
				fp.AddBTCDelDistInfo(&btcDel)
			}
		}

		// process all new BTC delegations under this finality provider
		if fpActiveBTCDels, ok := activeBTCDels[fpBTCPKHex]; ok {
			// handle new BTC delegations for this finality provider
			for _, d := range fpActiveBTCDels {
				fp.AddBTCDel(d)
			}
			// remove the finality provider entry in activeBTCDels map, so that
			// after the for loop the rest entries in activeBTCDels belongs to new
			// finality providers with new BTC delegations
			delete(activeBTCDels, fpBTCPKHex)
		}

		// add this finality provider to the new cache if it has voting power
		if fp.TotalVotingPower > 0 {
			newDc.AddFinalityProviderDistInfo(&fp)
		}
	}

	/*
		process new BTC delegations under new finality providers in activeBTCDels
	*/
	for fpBTCPKHex, fpActiveBTCDels := range activeBTCDels {
		// get the finality provider and initialise its dist info
		fpBTCPK, err := bbn.NewBIP340PubKeyFromHex(fpBTCPKHex)
		if err != nil {
			panic(err) // only programming error
		}
		newFP, err := k.GetFinalityProvider(ctx, *fpBTCPK)
		if err != nil {
			panic(err) // only programming error
		}
		fpDistInfo := types.NewFinalityProviderDistInfo(newFP)

		// add each BTC delegation
		for _, d := range fpActiveBTCDels {
			fpDistInfo.AddBTCDel(d)
		}

		// add this finality provider to the new cache if it has voting power
		if fpDistInfo.TotalVotingPower > 0 {
			newDc.AddFinalityProviderDistInfo(fpDistInfo)
		}
	}

	// get top N finality providers and their total voting power
	newDc.ApplyActiveFinalityProviders(k.GetParams(ctx).MaxActiveFinalityProviders)

	return newDc
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
func (k Keeper) GetAllPowerDistUpdateEvents(ctx context.Context, lastBTCTip uint64, curBTCTip uint64) []*types.EventPowerDistUpdate {
	events := []*types.EventPowerDistUpdate{}
	for i := lastBTCTip; i <= curBTCTip; i++ {
		k.IteratePowerDistUpdateEvents(ctx, i, func(event *types.EventPowerDistUpdate) bool {
			events = append(events, event)
			return true
		})
	}
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
