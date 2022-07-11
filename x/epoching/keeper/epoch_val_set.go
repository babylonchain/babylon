package keeper

import (
	"sort"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetValidatorSet returns the set of validators of a given epoch, where the validators are ordered by their voting power in descending order
func (k Keeper) GetValidatorSet(ctx sdk.Context, epochNumber uint64) []types.Validator {
	vals := types.ValidatorSet{}

	store := k.valSetStore(ctx, epochNumber)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Key())
		powerBytes := iterator.Value()
		var power sdk.Int
		if err := power.Unmarshal(powerBytes); err != nil {
			panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
		}
		val := types.Validator{
			Addr:  addr,
			Power: power.Int64(),
		}
		vals = append(vals, val)
	}
	sort.Sort(vals)

	return vals
}

// InitValidatorSet stores the validator set in the beginning of the current epoch
// This is called upon BeginBlock
func (k Keeper) InitValidatorSet(ctx sdk.Context) {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	store := k.valSetStore(ctx, epochNumber)
	totalPower := int64(0)

	// store the validator set
	k.stk.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		addrBytes := []byte(addr)
		powerBytes, err := sdk.NewInt(power).Marshal()
		if err != nil {
			panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
		}
		store.Set(addrBytes, powerBytes)
		totalPower += power

		return false
	})

	// store total voting power of this validator set
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	totalPowerBytes, err := sdk.NewInt(totalPower).Marshal()
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	k.votingPowerStore(ctx).Set(epochNumberBytes, totalPowerBytes)
}

// ClearValidatorSet removes the validator set of a given epoch
// TODO: This is called upon the epoch is checkpointed
func (k Keeper) ClearValidatorSet(ctx sdk.Context, epochNumber uint64) {
	store := k.valSetStore(ctx, epochNumber)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	// clear the validator set
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}
	// clear total voting power of this validator set
	powerStore := k.votingPowerStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	powerStore.Delete(epochNumberBytes)
}

// GetValidatorVotingPower returns the voting power of a given validator in a given epoch
func (k Keeper) GetValidatorVotingPower(ctx sdk.Context, epochNumber uint64, valAddr sdk.ValAddress) (int64, error) {
	store := k.valSetStore(ctx, epochNumber)

	powerBytes := store.Get(valAddr)
	if powerBytes == nil {
		return 0, types.ErrUnknownValidator
	}
	var power sdk.Int
	if err := power.Unmarshal(powerBytes); err != nil {
		panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
	}

	return power.Int64(), nil
}

// GetTotalVotingPower returns the total voting power of a given epoch
func (k Keeper) GetTotalVotingPower(ctx sdk.Context, epochNumber uint64) int64 {
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store := k.votingPowerStore(ctx)
	powerBytes := store.Get(epochNumberBytes)
	if powerBytes == nil {
		panic(types.ErrUnknownTotalVotingPower)
	}
	var power sdk.Int
	if err := power.Unmarshal(powerBytes); err != nil {
		panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
	}
	return power.Int64()
}

// valSetStore returns the KVStore of the validator set of a given epoch
// prefix: ValidatorSetKey || epochNumber
// key: string(address)
// value: voting power (in int64 as per Cosmos SDK)
func (k Keeper) valSetStore(ctx sdk.Context, epochNumber uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	valSetStore := prefix.NewStore(store, types.ValidatorSetKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	return prefix.NewStore(valSetStore, epochNumberBytes)
}

// votingPowerStore returns the total voting power of the validator set of a give nepoch
// prefix: ValidatorSetKey
// key: epochNumber
// value: total voting power (in int64 as per Cosmos SDK)
func (k Keeper) votingPowerStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.VotingPowerKey)
}
