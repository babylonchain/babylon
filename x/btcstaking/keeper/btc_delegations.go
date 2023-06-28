package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBTCDelegation adds the given BTC delegation to KVStore
func (k Keeper) SetBTCDelegation(ctx sdk.Context, btcDel *types.BTCDelegation) {
	store := k.btcDelegationStore(ctx, btcDel.ValBtcPk.MustMarshal())
	btcDelBytes := k.cdc.MustMarshal(btcDel)
	store.Set(btcDel.BtcPk.MustMarshal(), btcDelBytes)
}

// HasBTCDelegation checks if the given BTC delegation exists under a given BTC validator
func (k Keeper) HasBTCDelegation(ctx sdk.Context, valBTCPK []byte, delBTCPK []byte) bool {
	if !k.HasBTCValidator(ctx, valBTCPK) {
		return false
	}
	store := k.btcDelegationStore(ctx, valBTCPK)
	return store.Has(delBTCPK)
}

// GetBTCDelegation gets the BTC delegation with a given BTC PK under a given BTC validator
func (k Keeper) GetBTCDelegation(ctx sdk.Context, valBTCPK []byte, delBTCPK []byte) (*types.BTCDelegation, error) {
	// ensure the BTC validator exists
	if !k.HasBTCValidator(ctx, valBTCPK) {
		return nil, types.ErrBTCValNotFound
	}

	store := k.btcDelegationStore(ctx, valBTCPK)
	// ensure BTC delegation exists
	if !store.Has(delBTCPK) {
		return nil, types.ErrBTCDelNotFound
	}
	// get and unmarshal
	btcDelBytes := store.Get(delBTCPK)
	var btcDel types.BTCDelegation
	if err := k.cdc.Unmarshal(btcDelBytes, &btcDel); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BTC delegation: %w", err)
	}
	return &btcDel, nil
}

// btcDelegationStore returns the KVStore of the BTC delegations
// prefix: BTCDelegationKey || validator's Bitcoin secp256k1 PK
// key: delegation's Bitcoin secp256k1 PK
// value: BTCDelegation object
func (k Keeper) btcDelegationStore(ctx sdk.Context, valBTCPK []byte) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.BTCDelegationKey)
	return prefix.NewStore(delegationStore, valBTCPK)
}
