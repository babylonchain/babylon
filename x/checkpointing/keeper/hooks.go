package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements CheckpointingHooks interface
var _ types.CheckpointingHooks = Keeper{}

// AfterBlsKeyRegistered - call hook if registered
func (k Keeper) AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterBlsKeyRegistered(ctx, valAddr)
	}
	return nil
}

// AfterRawCheckpointSealed - call hook if the checkpoint is sealed
func (k Keeper) AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointSealed(ctx, epoch)
	}
	return nil
}

// AfterRawCheckpointConfirmed - call hook if the checkpoint is confirmed
func (k Keeper) AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointConfirmed(ctx, epoch)
	}
	return nil
}

func (k Keeper) AfterRawCheckpointForgotten(ctx context.Context, ckpt *types.RawCheckpoint) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointForgotten(ctx, ckpt)
	}
	return nil
}

// AfterRawCheckpointFinalized - call hook if the checkpoint is finalized
func (k Keeper) AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointFinalized(ctx, epoch)
	}
	return nil
}

// AfterRawCheckpointBlsSigVerified - call hook if the checkpoint's BLS sig is verified
func (k Keeper) AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *types.RawCheckpoint) error {
	if k.hooks != nil {
		return k.hooks.AfterRawCheckpointBlsSigVerified(ctx, ckpt)
	}
	return nil
}
