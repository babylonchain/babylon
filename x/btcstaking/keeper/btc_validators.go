package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBTCValidator adds the given BTC validator to KVStore
func (k Keeper) SetBTCValidator(ctx sdk.Context, btcVal *types.BTCValidator) {
	store := k.btcValidatorStore(ctx)
	btcValBytes := k.cdc.MustMarshal(btcVal)
	store.Set(btcVal.BtcPk.MustMarshal(), btcValBytes)
}

// HasBTCValidator checks if the BTC validator exists
func (k Keeper) HasBTCValidator(ctx sdk.Context, valBTCPK []byte) bool {
	store := k.btcValidatorStore(ctx)
	return store.Has(valBTCPK)
}

// GetBTCValidator gets the BTC validator with the given validator Bitcoin PK
func (k Keeper) GetBTCValidator(ctx sdk.Context, valBTCPK []byte) (*types.BTCValidator, error) {
	store := k.btcValidatorStore(ctx)
	if !k.HasBTCValidator(ctx, valBTCPK) {
		return nil, types.ErrBTCValNotFound
	}
	btcValBytes := store.Get(valBTCPK)
	var btcVal types.BTCValidator
	if err := k.cdc.Unmarshal(btcValBytes, &btcVal); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BTC validator: %w", err)
	}
	return &btcVal, nil
}

// btcValidatorStore returns the KVStore of the BTC validator set
// prefix: BTCValidatorKey
// key: Bitcoin secp256k1 PK
// value: BTCValidator object
func (k Keeper) btcValidatorStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.BTCValidatorKey)
}
