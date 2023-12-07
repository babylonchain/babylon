package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

// SetBTCValidator adds the given BTC validator to KVStore
func (k Keeper) SetBTCValidator(ctx context.Context, btcVal *types.BTCValidator) {
	store := k.btcValidatorStore(ctx)
	btcValBytes := k.cdc.MustMarshal(btcVal)
	store.Set(btcVal.BtcPk.MustMarshal(), btcValBytes)
}

// HasBTCValidator checks if the BTC validator exists
func (k Keeper) HasBTCValidator(ctx context.Context, valBTCPK []byte) bool {
	store := k.btcValidatorStore(ctx)
	return store.Has(valBTCPK)
}

// GetBTCValidator gets the BTC validator with the given validator Bitcoin PK
func (k Keeper) GetBTCValidator(ctx context.Context, valBTCPK []byte) (*types.BTCValidator, error) {
	store := k.btcValidatorStore(ctx)
	if !k.HasBTCValidator(ctx, valBTCPK) {
		return nil, types.ErrBTCValNotFound
	}
	btcValBytes := store.Get(valBTCPK)
	var btcVal types.BTCValidator
	k.cdc.MustUnmarshal(btcValBytes, &btcVal)
	return &btcVal, nil
}

// SlashBTCValidator slashes a BTC validator with the given PK
// A slashed BTC validator will not have voting power
func (k Keeper) SlashBTCValidator(ctx context.Context, valBTCPK []byte) error {
	btcVal, err := k.GetBTCValidator(ctx, valBTCPK)
	if err != nil {
		return err
	}
	if btcVal.IsSlashed() {
		return types.ErrBTCValAlreadySlashed
	}
	btcVal.SlashedBabylonHeight = uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	btcTip := k.btclcKeeper.GetTipInfo(ctx)
	if btcTip == nil {
		panic(fmt.Errorf("failed to get current BTC tip"))
	}
	btcVal.SlashedBtcHeight = btcTip.Height
	k.SetBTCValidator(ctx, btcVal)
	return nil
}

// btcValidatorStore returns the KVStore of the BTC validator set
// prefix: BTCValidatorKey
// key: Bitcoin secp256k1 PK
// value: BTCValidator object
func (k Keeper) btcValidatorStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.BTCValidatorKey)
}
