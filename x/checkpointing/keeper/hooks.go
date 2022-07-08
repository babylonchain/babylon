package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements CheckpointingHooks interface
var _ types.CheckpointingHooks = Keeper{}

// AfterBlsKeyRegistered - call hook if registered
func (k Keeper) AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterBlsKeyRegistered(ctx, valAddr)
	}
	return nil
}

// AfterRawCheckpointConfirmed - call hook if registered
func (k Keeper) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointConfirmed(ctx, epoch)
	}
	return nil
}
