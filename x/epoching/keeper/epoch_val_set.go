package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetValidatorSet returns the set of validators of a given epoch, where the validators are ordered by their address in ascending order
func (k Keeper) GetValidatorSet(ctx context.Context, epochNumber uint64) types.ValidatorSet {
	vals := []types.Validator{}

	store := k.valSetStore(ctx, epochNumber)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.ValAddress(iterator.Key())
		powerBytes := iterator.Value()
		var power sdkmath.Int
		if err := power.Unmarshal(powerBytes); err != nil {
			panic(errorsmod.Wrap(types.ErrUnmarshal, err.Error()))
		}
		val := types.Validator{
			Addr:  addr,
			Power: power.Int64(),
		}
		vals = append(vals, val)
	}
	return types.NewSortedValidatorSet(vals)
}

func (k Keeper) GetCurrentValidatorSet(ctx context.Context) types.ValidatorSet {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	return k.GetValidatorSet(ctx, epochNumber)
}

// InitValidatorSet stores the validator set in the beginning of the current epoch
// This is called upon BeginBlock
func (k Keeper) InitValidatorSet(ctx context.Context) {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	store := k.valSetStore(ctx, epochNumber)
	totalPower := int64(0)

	// store the validator set
	err := k.stk.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		addrBytes := []byte(addr)
		powerBytes, err := sdkmath.NewInt(power).Marshal()
		if err != nil {
			panic(errorsmod.Wrap(types.ErrMarshal, err.Error()))
		}
		store.Set(addrBytes, powerBytes)
		totalPower += power

		return false
	})

	if err != nil {
		panic(err)
	}

	// store total voting power of this validator set
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	totalPowerBytes, err := sdkmath.NewInt(totalPower).Marshal()
	if err != nil {
		panic(errorsmod.Wrap(types.ErrMarshal, err.Error()))
	}
	k.votingPowerStore(ctx).Set(epochNumberBytes, totalPowerBytes)
}

// ClearValidatorSet removes the validator set of a given epoch
// TODO: This is called upon the epoch is checkpointed
func (k Keeper) ClearValidatorSet(ctx context.Context, epochNumber uint64) {
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
func (k Keeper) GetValidatorVotingPower(ctx context.Context, epochNumber uint64, valAddr sdk.ValAddress) (int64, error) {
	store := k.valSetStore(ctx, epochNumber)

	powerBytes := store.Get(valAddr)
	if powerBytes == nil {
		return 0, types.ErrUnknownValidator
	}
	var power sdkmath.Int
	if err := power.Unmarshal(powerBytes); err != nil {
		panic(errorsmod.Wrap(types.ErrUnmarshal, err.Error()))
	}

	return power.Int64(), nil
}

func (k Keeper) GetCurrentValidatorVotingPower(ctx context.Context, valAddr sdk.ValAddress) (int64, error) {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	return k.GetValidatorVotingPower(ctx, epochNumber, valAddr)
}

// GetTotalVotingPower returns the total voting power of a given epoch
func (k Keeper) GetTotalVotingPower(ctx context.Context, epochNumber uint64) int64 {
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store := k.votingPowerStore(ctx)
	powerBytes := store.Get(epochNumberBytes)
	if powerBytes == nil {
		panic(types.ErrUnknownTotalVotingPower)
	}
	var power sdkmath.Int
	if err := power.Unmarshal(powerBytes); err != nil {
		panic(errorsmod.Wrap(types.ErrUnmarshal, err.Error()))
	}
	return power.Int64()
}

// valSetStore returns the KVStore of the validator set of a given epoch
// prefix: ValidatorSetKey || epochNumber
// key: string(address)
// value: voting power (in int64 as per Cosmos SDK)
func (k Keeper) valSetStore(ctx context.Context, epochNumber uint64) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	valSetStore := prefix.NewStore(storeAdapter, types.ValidatorSetKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	return prefix.NewStore(valSetStore, epochNumberBytes)
}

// votingPowerStore returns the total voting power of the validator set of a given epoch
// prefix: ValidatorSetKey
// key: epochNumber
// value: total voting power (in int64 as per Cosmos SDK)
func (k Keeper) votingPowerStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.VotingPowerKey)
}
