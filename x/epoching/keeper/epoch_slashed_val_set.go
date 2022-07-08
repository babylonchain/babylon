package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// setSlashedVotingPower sets the total amount of voting power that has been slashed in the epoch
func (k Keeper) setSlashedVotingPower(ctx sdk.Context, epochNumber sdk.Uint, power int64) {
	store := k.slashedVotingPowerStore(ctx)

	// key: epochNumber
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	// value: power
	powerBytes, err := sdk.NewInt(power).Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}

	store.Set(epochNumberBytes, powerBytes)
}

// InitSlashedVotingPower sets the slashed voting power of the current epoch to 0
// This is called upon initialising the genesis state and upon a new epoch
func (k Keeper) InitSlashedVotingPower(ctx sdk.Context) {
	epochNumber := k.GetEpochNumber(ctx)
	k.setSlashedVotingPower(ctx, epochNumber, 0)
}

// GetSlashedVotingPower fetches the amount of slashed voting power of a given epoch
func (k Keeper) GetSlashedVotingPower(ctx sdk.Context, epochNumber sdk.Uint) int64 {
	store := k.slashedVotingPowerStore(ctx)

	// key: epochNumber
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	bz := store.Get(epochNumberBytes)
	if bz == nil {
		panic(types.ErrUnknownSlashedVotingPower)
	}
	// get value
	var slashedVotingPower sdk.Int
	if err := slashedVotingPower.Unmarshal(bz); err != nil {
		panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
	}

	return slashedVotingPower.Int64()
}

// AddSlashedValidator adds a slashed validator to the set of the current epoch
// This is called upon hook `BeforeValidatorSlashed` exposed by the staking module
func (k Keeper) AddSlashedValidator(ctx sdk.Context, valAddr sdk.ValAddress) {
	epochNumber := k.GetEpochNumber(ctx)
	store := k.slashedValSetStore(ctx, epochNumber)

	// insert into "set of slashed addresses" as KV pair, where
	// - key: valAddr
	// - value: empty
	store.Set(valAddr, []byte{})

	// add voting power
	slashedVotingPower := k.GetSlashedVotingPower(ctx, epochNumber)
	thisVotingPower := k.GetValidatorVotingPower(ctx, epochNumber, valAddr)
	k.setSlashedVotingPower(ctx, epochNumber, slashedVotingPower+thisVotingPower)
}

// GetSlashedValidators returns the set of slashed validators of a given epoch
func (k Keeper) GetSlashedValidators(ctx sdk.Context, epochNumber sdk.Uint) []sdk.ValAddress {
	addrs := []sdk.ValAddress{}
	store := k.slashedValSetStore(ctx, epochNumber)
	// add each valAddr, which is the key
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Key())
		addrs = append(addrs, addr)
	}

	return addrs
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

	// forget the slashed voting power of this epoch
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	k.slashedVotingPowerStore(ctx).Delete(epochNumberBytes)
}

// slashedValSetStore returns the KVStore of the slashed validator set for a given epoch
// prefix : SlashedValidatorSetKey || epochNumber
func (k Keeper) slashedValSetStore(ctx sdk.Context, epochNumber sdk.Uint) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	slashedValStore := prefix.NewStore(store, types.SlashedValidatorSetKey)
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	return prefix.NewStore(slashedValStore, epochNumberBytes)
}

// slashedVotingPower returns the KVStore of the slashed voting power
// prefix: SlashedVotingPowerKey
func (k Keeper) slashedVotingPowerStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.SlashedVotingPowerKey)
}
