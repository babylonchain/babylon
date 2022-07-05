package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSlashedValidatorSetSize fetches the number of slashed validators
func (k Keeper) GetSlashedValidatorSetSize(ctx sdk.Context) sdk.Uint {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.SlashedValidatorSetSizeKey)
	if bz == nil {
		panic(types.ErrUnknownSlashedValSetSize)
	}
	var len sdk.Uint
	if err := len.Unmarshal(bz); err != nil {
		panic(err)
	}

	return len
}

// SetSlashedValidatorSetSize sets the slashed validator set size
func (k Keeper) SetSlashedValidatorSetSize(ctx sdk.Context, len sdk.Uint) {
	store := ctx.KVStore(k.storeKey)

	lenBytes, err := len.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.SlashedValidatorSetSizeKey, lenBytes)
}

// incSlashedValidatorSetSize adds the set size by 1
func (k Keeper) incSlashedValidatorSetSize(ctx sdk.Context) sdk.Uint {
	len := k.GetSlashedValidatorSetSize(ctx)
	incLen := len.AddUint64(1)
	k.SetSlashedValidatorSetSize(ctx, incLen)
	return incLen
}

// AddSlashedValidator adds a slashed validator to the set of the current epoch
func (k Keeper) AddSlashedValidator(ctx sdk.Context, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)

	// insert KV pair, where
	// - key: SlashedValidatorKey || lenBytes
	// - value: valAddr in Bytes
	len := k.GetSlashedValidatorSetSize(ctx)
	lenBytes, err := len.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(append(types.SlashedValidatorKey, lenBytes...), valAddr)

	// increment set size
	k.incSlashedValidatorSetSize(ctx)
}

// GetSlashedValidators returns the set of slashed validators in the current epoch
func (k Keeper) GetSlashedValidators(ctx sdk.Context) []sdk.ValAddress {
	addrs := []sdk.ValAddress{}
	store := ctx.KVStore(k.storeKey)

	// add each slashed validator addr to the set
	iterator := sdk.KVStorePrefixIterator(store, types.SlashedValidatorKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Value())
		addrs = append(addrs, addr)
	}

	return addrs
}

// ClearSlashedValidators removes all slashed validators in the set
func (k Keeper) ClearSlashedValidators(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// remove all slashed validators
	iterator := sdk.KVStorePrefixIterator(store, types.SlashedValidatorKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set set size to zero
	k.SetSlashedValidatorSetSize(ctx, sdk.NewUint(0))
}
