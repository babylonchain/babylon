package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements BTCLightClientHooks interface
var _ types.BTCLightClientHooks = Keeper{}

// AfterTipUpdated - call hook if registered
func (k Keeper) AfterTipUpdated(ctx sdk.Context, height uint64) {
	if k.hooks != nil {
		k.hooks.AfterTipUpdated(ctx, height)
	}
}
