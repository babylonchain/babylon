package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetValidatorSet returns the set of validators of a given epoch
func (k Keeper) GetValidatorSet(ctx sdk.Context, epochNumber sdk.Uint) map[string]int64 {
	valSet := make(map[string]int64)
	store := k.valSetStore(ctx, epochNumber)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := string(iterator.Key())
		powerBytes := iterator.Value()
		var power sdk.Int
		if err := power.Unmarshal(powerBytes); err != nil {
			panic(err)
		}
		valSet[addr] = power.Int64()
	}

	return valSet
}

// GetTotalVotingPower returns the total voting power of a given epoch
func (k Keeper) GetTotalVotingPower(ctx sdk.Context, epochNumber sdk.Uint) int64 {
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	store := k.votingPowerStore(ctx)
	powerBytes := store.Get(epochNumberBytes)
	var power sdk.Int
	if err := power.Unmarshal(powerBytes); err != nil {
		panic(err)
	}
	return power.Int64()
}

// SetValidatorSet stores the validator set in the beginning of a given epoch
// This is called upon BeginBlock
func (k Keeper) SetValidatorSet(ctx sdk.Context, epochNumber sdk.Uint) {
	store := k.valSetStore(ctx, epochNumber)
	totalPower := int64(0)

	// store the validator set
	valSet, err := k.getValSetFromStaking(ctx)
	if err != nil {
		panic(err)
	}
	for addr, power := range valSet {
		addrBytes := []byte(addr)
		powerBytes, err := sdk.NewInt(power).Marshal()
		if err != nil {
			panic(err)
		}
		store.Set(addrBytes, powerBytes)
		totalPower += power
	}
	// store total voting power of this validator set
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	totalPowerBytes, err := sdk.NewInt(totalPower).Marshal()
	if err != nil {
		panic(err)
	}
	k.votingPowerStore(ctx).Set(epochNumberBytes, totalPowerBytes)
}

// ClearValidatorSet removes the validator set of a given epoch
// TODO: This is called upon the epoch is checkpointed
func (k Keeper) ClearValidatorSet(ctx sdk.Context, epochNumber sdk.Uint) {
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
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
	powerStore.Delete(epochNumberBytes)
}

// valSetStore returns the KVStore of the validator set of a given epoch
// prefix: ValidatorSetKey || epochNumber
// key: string(address)
// value: voting power (in int64 as per Cosmos SDK)
func (k Keeper) valSetStore(ctx sdk.Context, epochNumber sdk.Uint) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	valSetStore := prefix.NewStore(store, types.ValidatorSetKey)
	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}
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

// get the last validator set
// key: string(address)
// value: voting power (in int64 as per Cosmos SDK)
// This is called upon BeginEpoch
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/staking/keeper/val_state_change.go#L348-L373)
func (k Keeper) getValSetFromStaking(ctx sdk.Context) (map[string]int64, error) {
	valSet := make(map[string]int64)

	iterator := k.stk.LastValidatorsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// extract the validator address from the key (prefix is 1-byte, addrLen is 1-byte)
		valAddr := stakingtypes.AddressFromLastValidatorPowerKey(iterator.Key())
		valAddrStr := string(valAddr)
		powerBytes := iterator.Value()
		var power sdk.Int
		if err := power.Unmarshal(powerBytes); err != nil {
			return nil, err
		}
		valSet[valAddrStr] = power.Int64()
	}

	return valSet, nil
}
