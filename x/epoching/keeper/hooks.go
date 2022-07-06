package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements EpochingHooks interface
var _ types.EpochingHooks = Keeper{}

// AfterEpochBegins - call hook if registered
func (k Keeper) AfterEpochBegins(ctx sdk.Context, epoch sdk.Uint) {
	if k.hooks != nil {
		k.hooks.AfterEpochBegins(ctx, epoch)
	}
}

// AfterEpochEnds - call hook if registered
func (k Keeper) AfterEpochEnds(ctx sdk.Context, epoch sdk.Uint) {
	if k.hooks != nil {
		k.hooks.AfterEpochEnds(ctx, epoch)
	}
}
