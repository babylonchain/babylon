package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements EpochingHooks interface
var _ types.EpochingHooks = Keeper{}

// BeginEpoch - call hook if registered
func (k Keeper) BeginEpoch(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.BeginEpoch(ctx, epoch)
	}
	return nil
}

// EndEpoch - call hook if registered
func (k Keeper) EndEpoch(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.BeginEpoch(ctx, epoch)
	}
	return nil
}
