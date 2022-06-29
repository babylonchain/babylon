package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSlashedValidatorSetSize fetches the number of slashed validators
func (k Keeper) GetSlashedValidatorSetSize(ctx sdk.Context) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.SlashedValidatorSetSizeKey)
	if bz == nil {
		return sdk.NewUint(0), nil
	}
	var len sdk.Uint
	err := len.Unmarshal(bz)

	return len, err
}

// setSlashedValidatorSetSize sets the slashed validator set size
func (k Keeper) setSlashedValidatorSetSize(ctx sdk.Context, len sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)

	lenBytes, err := len.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.SlashedValidatorSetSizeKey, lenBytes)

	return nil
}

// incSlashedValidatorSetSize adds the set size by 1
func (k Keeper) incSlashedValidatorSetSize(ctx sdk.Context) error {
	len, err := k.GetSlashedValidatorSetSize(ctx)
	if err != nil {
		return err
	}
	incLen := len.AddUint64(1)
	return k.setSlashedValidatorSetSize(ctx, incLen)
}

// AddSlashedValidator adds a slashed validator to the set of the current epoch
func (k Keeper) AddSlashedValidator(ctx sdk.Context, valAddr sdk.ValAddress) error {
	store := ctx.KVStore(k.storeKey)

	// insert KV pair, where
	// - key: SlashedValidatorKey || lenBytes
	// - value: valAddr in Bytes
	len, err := k.GetSlashedValidatorSetSize(ctx)
	if err != nil {
		return err
	}
	lenBytes, err := len.Marshal()
	if err != nil {
		return err
	}
	store.Set(append(types.SlashedValidatorKey, lenBytes...), valAddr)

	// increment set size
	return k.incSlashedValidatorSetSize(ctx)
}

// GetSlashedValidators returns the set of slashed validators in the current epoch
func (k Keeper) GetSlashedValidators(ctx sdk.Context) ([]sdk.ValAddress, error) {
	addrs := []sdk.ValAddress{}
	store := ctx.KVStore(k.storeKey)

	// add each slashed validator addr to the set
	iterator := sdk.KVStorePrefixIterator(store, types.SlashedValidatorKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Value())
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

// ClearSlashedValidators removes all slashed validators in the set
func (k Keeper) ClearSlashedValidators(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	// remove all slashed validators
	iterator := sdk.KVStorePrefixIterator(store, types.SlashedValidatorKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set set size to zero
	return k.setSlashedValidatorSetSize(ctx, sdk.NewUint(0))
}

// TODO: upon EndBlock, clear the set of slashed validators
