package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

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
	fp, err := k.GetFinalityProvider(ctx, fpBTCPK)
	if err != nil {
		return err
	}
	if fp.IsSlashed() {
		return types.ErrFpAlreadySlashed
	}
	fp.SlashedBabylonHeight = uint64(sdk.UnwrapSDKContext(ctx).BlockHeight())
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	if btcTip == nil {
		panic(fmt.Errorf("failed to get current BTC tip"))
	}
	fp.SlashedBtcHeight = btcTip.Height
	k.SetFinalityProvider(ctx, fp)
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
