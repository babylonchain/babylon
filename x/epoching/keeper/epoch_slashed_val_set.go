package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setSlashedValidatorSetSize sets the slashed validator set size
func (k Keeper) setSlashedValidatorSetSize(ctx sdk.Context, epochNumber sdk.Uint, size sdk.Uint) {
	store := k.slashedValSetSizeStore(ctx)

	// key: epochNumber
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	// value: setSize
	sizeBytes, err := size.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(epochNumberBytes, sizeBytes)
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
	epochNumber := k.GetEpochNumber(ctx)
	store := k.slashedValSetStore(ctx, epochNumber)

	// insert KV pair, where
	// - key: valAddr
	// - value: empty
	store.Set(valAddr, []byte{})

	// increment set size
	size := k.GetSlashedValidatorSetSize(ctx, epochNumber)
	incSize := size.AddUint64(1)
	k.setSlashedValidatorSetSize(ctx, epochNumber, incSize)
}

// GetSlashedValidators returns the set of slashed validators of a given epoch
func (k Keeper) GetSlashedValidators(ctx sdk.Context, epochNumber sdk.Uint) []sdk.ValAddress {
	addrs := []sdk.ValAddress{}
	store := k.slashedValSetStore(ctx, epochNumber)
	// add each valAddr, which is the key
	// the prefix `SlashedValidatorSetKey || epochNumber` has been stripped
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Key())
		addrs = append(addrs, addr)
	}

	return addrs
}

// GetSlashedValidatorSetSize fetches the number of slashed validators of a given epoch
func (k Keeper) GetSlashedValidatorSetSize(ctx sdk.Context, epochNumber sdk.Uint) sdk.Uint {
	// prefix: SlashedValidatorSetSizeKey
	store := k.slashedValSetSizeStore(ctx)

	// key: epochNumber
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	bz := store.Get(epochNumberBytes)
	if bz == nil {
		panic(types.ErrUnknownSlashedValSetSize)
	}
	var setSize sdk.Uint
	if err := setSize.Unmarshal(bz); err != nil {
		panic(err)
	}

	return setSize
}

// ClearSlashedValidators removes all slashed validators in the set
// TODO: This is called upon the epoch is checkpointed
func (k Keeper) ClearSlashedValidators(ctx sdk.Context, epochNumber sdk.Uint) {
	// prefix : SlashedValidatorSetKey || epochNumber
	store := k.slashedValSetStore(ctx, epochNumber)

	// remove all entries with this prefix
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set the set size of this epoch to zero
	k.setSlashedValidatorSetSize(ctx, epochNumber, sdk.NewUint(0))
}

// slashedValSetStore returns the KVStore of the slashed validator set for a given epoch
// prefix : SlashedValidatorSetKey || epochNumber
func (k Keeper) slashedValSetStore(ctx sdk.Context, epochNumber sdk.Uint) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	slashedValStore := prefix.NewStore(store, types.SlashedValidatorSetKey)
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	return prefix.NewStore(slashedValStore, epochNumberBytes)
}

// slashedValSetSizeStore returns the KVStore of the slashed validator set size
// prefix: SlashedValidatorSetSizeKey
func (k Keeper) slashedValSetSizeStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.SlashedValidatorSetSizeKey)
}
