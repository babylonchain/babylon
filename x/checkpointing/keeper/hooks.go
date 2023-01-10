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

// AfterRawCheckpointSubmitted - call hook if the checkpoint is sealed
func (k Keeper) AfterRawCheckpointSealed(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointSealed(ctx, epoch)
	}
	return nil
}

// AfterRawCheckpointSubmitted - call hook if the checkpoint is submitted
func (k Keeper) AfterRawCheckpointSubmitted(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointSubmitted(ctx, epoch)
	}
	return nil
}

// AfterRawCheckpointConfirmed - call hook if the checkpoint is confirmed
func (k Keeper) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointConfirmed(ctx, epoch)
	}
	return nil
}

// AfterRawCheckpointFinalized - call hook if the checkpoint is finalized
func (k Keeper) AfterRawCheckpointFinalized(ctx sdk.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointFinalized(ctx, epoch)
	}
	return nil
}
