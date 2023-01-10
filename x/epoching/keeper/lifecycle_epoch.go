package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RecordNewEpochState adds a state for an epoch lifecycle, including
func (k Keeper) RecordNewEpochState(ctx sdk.Context, epochNumber uint64, state types.EpochState) error {
	lc := k.GetEpochLifecycle(ctx, epochNumber)
	if lc == nil {
		lc = &types.EpochLifecycle{
			EpochNumber: epochNumber,
			EpochLife:   []*types.EpochStateUpdate{},
		}
	}
	height, time := ctx.BlockHeight(), ctx.BlockTime()
	epochStateUpdate := types.EpochStateUpdate{
		State:       state,
		BlockHeight: uint64(height),
		BlockTime:   &time,
	}
	lc.EpochLife = append(lc.EpochLife, &epochStateUpdate)
	k.SetEpochLifecycle(ctx, epochNumber, lc)
	return nil
}

func (k Keeper) SetEpochLifecycle(ctx sdk.Context, epochNumber uint64, lc *types.EpochLifecycle) {
	store := k.epochLifecycleStore(ctx)
	lcBytes := k.cdc.MustMarshal(lc)
	store.Set(sdk.Uint64ToBigEndian(epochNumber), lcBytes)
}

func (k Keeper) GetEpochLifecycle(ctx sdk.Context, epochNumber uint64) *types.EpochLifecycle {
	store := k.epochLifecycleStore(ctx)
	lcBytes := store.Get(sdk.Uint64ToBigEndian(epochNumber))
	if len(lcBytes) == 0 {
		return nil
	}
	var lc types.EpochLifecycle
	k.cdc.MustUnmarshal(lcBytes, &lc)
	return &lc
}

// epochLifecycleStore returns the store of the epoch lifecycle
// prefix: EpochLifecycleKey
// key: epoch_number
// value: EpochLifecycle object
func (k Keeper) epochLifecycleStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.EpochLifecycleKey)
}
