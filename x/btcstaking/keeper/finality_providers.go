package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

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

/* finality provider storage */

// SetFinalityProvider adds the given finality provider to KVStore
func (k Keeper) SetFinalityProvider(ctx context.Context, fp *types.FinalityProvider) {
	store := k.finalityProviderStore(ctx)
	fpBytes := k.cdc.MustMarshal(fp)
	store.Set(fp.BtcPk.MustMarshal(), fpBytes)
}

// HasFinalityProvider checks if the finality provider exists
func (k Keeper) HasFinalityProvider(ctx context.Context, fpBTCPK []byte) bool {
	store := k.finalityProviderStore(ctx)
	return store.Has(fpBTCPK)
}

// GetFinalityProvider gets the finality provider with the given finality provider Bitcoin PK
func (k Keeper) GetFinalityProvider(ctx context.Context, fpBTCPK []byte) (*types.FinalityProvider, error) {
	store := k.finalityProviderStore(ctx)
	if !k.HasFinalityProvider(ctx, fpBTCPK) {
		return nil, types.ErrFpNotFound
	}
	fpBytes := store.Get(fpBTCPK)
	var fp types.FinalityProvider
	k.cdc.MustUnmarshal(fpBytes, &fp)
	return &fp, nil
}

// SlashFinalityProvider slashes a finality provider with the given PK
// A slashed finality provider will not have voting power
func (k Keeper) SlashFinalityProvider(ctx context.Context, fpBTCPK []byte) error {
	// ensure finality provider exists
	fp, err := k.GetFinalityProvider(ctx, fpBTCPK)
	if err != nil {
		return err
	}

	// ensure finality provider is not slashed yet
	if fp.IsSlashed() {
		return types.ErrFpAlreadySlashed
	}

	// set finality provider to be slashed
	fp.SlashedBabylonHeight = uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	if btcTip == nil {
		panic(fmt.Errorf("failed to get current BTC tip"))
	}
	fp.SlashedBtcHeight = btcTip.Height
	k.SetFinalityProvider(ctx, fp)

	// record slashed event
	k.setFinalityProviderEvent(ctx, fpBTCPK)

	return nil
}

// finalityProviderStore returns the KVStore of the finality provider set
// prefix: FinalityProviderKey
// key: Bitcoin secp256k1 PK
// value: FinalityProvider object
func (k Keeper) finalityProviderStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.FinalityProviderKey)
}
