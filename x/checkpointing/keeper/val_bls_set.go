package keeper

import (
	"fmt"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetValidatorBlsKeySet returns the set of validators of a given epoch with BLS public key
// the validators are ordered by their address in ascending order
func (k Keeper) GetValidatorBlsKeySet(ctx sdk.Context, epochNumber uint64) *types.ValidatorWithBlsKeySet {
	store := k.valBlsSetStore(ctx, epochNumber)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	valBlsKeySetBytes := store.Get(epochNumberBytes)
	valBlsKeySet, err := types.BytesToValidatorBlsKeySet(k.cdc, valBlsKeySetBytes)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal validator BLS key set: %w", err))
	}
	return valBlsKeySet
}

func (k Keeper) GetCurrentValidatorBlsKeySet(ctx sdk.Context) *types.ValidatorWithBlsKeySet {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	return k.GetValidatorBlsKeySet(ctx, epochNumber)
}

// InitValidatorBLSSet stores the validator set with BLS keys in the beginning of the current epoch
// This is called upon BeginBlock
func (k Keeper) InitValidatorBLSSet(ctx sdk.Context) error {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	valset := k.GetValidatorSet(ctx, epochNumber)
	valBlsSet := &types.ValidatorWithBlsKeySet{
		ValSet: make([]*types.ValidatorWithBlsKey, len(valset)),
	}
	for i, val := range valset {
		blsPubkey, err := k.GetBlsPubKey(ctx, val.Addr)
		if err != nil {
			return fmt.Errorf("failed to get BLS public key of address %v: %w", val.Addr, err)
		}
		valBls := &types.ValidatorWithBlsKey{
			ValidatorAddress: val.GetValAddressStr(),
			BlsPubKey:        blsPubkey,
			VotingPower:      uint64(val.Power),
		}
		valBlsSet.ValSet[i] = valBls
	}
	valBlsSetBytes := types.ValidatorBlsKeySetToBytes(k.cdc, valBlsSet)
	store := k.valBlsSetStore(ctx, epochNumber)
	store.Set(types.ValidatorBlsKeySetKey(epochNumber), valBlsSetBytes)

	return nil
}

// ClearValidatorSet removes the validator BLS set of a given epoch
// TODO: This is called upon the epoch is checkpointed
func (k Keeper) ClearValidatorSet(ctx sdk.Context, epochNumber uint64) {
	store := k.valBlsSetStore(ctx, epochNumber)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Delete(epochNumberBytes)
}

// valBlsSetStore returns the KVStore of the validator BLS set of a given epoch
// prefix: ValidatorBLSSetKey
// key: epoch number
// value: ValidatorBLSKeySet
func (k Keeper) valBlsSetStore(ctx sdk.Context, epochNumber uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.ValidatorBlsKeySetPrefix)
}
