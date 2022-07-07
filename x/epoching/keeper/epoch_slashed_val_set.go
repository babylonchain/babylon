package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setSlashedValidatorSetSize sets the slashed validator set size
func (k Keeper) setSlashedValidatorSetSize(ctx sdk.Context, epochNumber sdk.Uint, size sdk.Uint) {
	store := ctx.KVStore(k.storeKey)

	// key: SlashedValidatorSetSizeKey || epochNumber
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	key := append(types.SlashedValidatorSetSizeKey, epochNumberBytes...)

	sizeBytes, err := size.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(key, sizeBytes)
}

// InitSlashedValidatorSetSize sets the slashed validator set size of the current epoch to 0
// This is called upon initialising the genesis state and upon a new epoch
func (k Keeper) InitSlashedValidatorSetSize(ctx sdk.Context) {
	epochNumber := k.GetEpochNumber(ctx)
	k.setSlashedValidatorSetSize(ctx, epochNumber, sdk.NewUint(0))
}

// AddSlashedValidator adds a slashed validator to the set of the current epoch
// This is called upon hook `BeforeValidatorSlashed` exposed by the staking module
func (k Keeper) AddSlashedValidator(ctx sdk.Context, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)

	// insert KV pair, where
	// - key: SlashedValidatorKey || epochNumber || valAddr
	// - value: empty
	epochNumber := k.GetEpochNumber(ctx)
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	key := types.SlashedValidatorKey
	key = append(key, epochNumberBytes...)
	key = append(key, valAddr...)
	store.Set(key, []byte{})

	// increment set size
	size := k.GetSlashedValidatorSetSize(ctx, epochNumber)
	incSize := size.AddUint64(1)
	k.setSlashedValidatorSetSize(ctx, epochNumber, incSize)
}

// GetSlashedValidators returns the set of slashed validators of a given epoch
func (k Keeper) GetSlashedValidators(ctx sdk.Context, epochNumber sdk.Uint) []sdk.ValAddress {
	addrs := []sdk.ValAddress{}
	store := ctx.KVStore(k.storeKey)

	// add each slashed validator addr to the set
	// key: SlashedValidatorKey || epochNumber || valAddr
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	prefix := append(types.SlashedValidatorKey, epochNumberBytes...)

	// add each valAddr, which is the key
	// the prefix `SlashedValidatorKey || epochNumber` has been excluded
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Key())
		addrs = append(addrs, addr)
	}

	return addrs
}

// GetSlashedValidatorSetSize fetches the number of slashed validators of a given epoch
func (k Keeper) GetSlashedValidatorSetSize(ctx sdk.Context, epochNumber sdk.Uint) sdk.Uint {
	store := ctx.KVStore(k.storeKey)

	// key: SlashedValidatorSetSizeKey || epochNumber
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	key := append(types.SlashedValidatorSetSizeKey, epochNumberBytes...)

	bz := store.Get(key)
	if bz == nil {
		panic(types.ErrUnknownSlashedValSetSize)
	}
	var size sdk.Uint
	if err := size.Unmarshal(bz); err != nil {
		panic(err)
	}

	return size
}

// ClearSlashedValidators removes all slashed validators in the set
// This is called upon the epoch is checkpointed
func (k Keeper) ClearSlashedValidators(ctx sdk.Context, epochNumber sdk.Uint) {
	store := ctx.KVStore(k.storeKey)

	// key: SlashedValidatorKey || epochNumber || valAddr
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	prefix := append(types.SlashedValidatorKey, epochNumberBytes...)

	// remove all entries with prefix SlashedValidatorKey || epochNumber
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set the set size of this epoch to zero
	k.setSlashedValidatorSetSize(ctx, epochNumber, sdk.NewUint(0))
}
