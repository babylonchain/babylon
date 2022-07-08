package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements BTCLightClientHooks interface
var _ types.BTCLightClientHooks = Keeper{}

// AfterBTCRollBack - call hook if registered
func (k Keeper) AfterBTCRollBack(ctx sdk.Context, hash bbl.BTCHeaderHashBytes, height uint64) {
	if k.hooks != nil {
		k.hooks.AfterBTCRollBack(ctx, hash, height)
	}
}

// AfterBTCRollForward - call hook if registered
func (k Keeper) AfterBTCRollForward(ctx sdk.Context, hash bbl.BTCHeaderHashBytes, height uint64) {
	if k.hooks != nil {
		k.hooks.AfterBTCRollBack(ctx, hash, height)
	}
}
