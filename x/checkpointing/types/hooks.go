package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple Checkpointing hooks, all hook functions are run in array sequence
var _ CheckpointingHooks = &MultiCheckpointingHooks{}

type MultiCheckpointingHooks []CheckpointingHooks

func NewMultiCheckpointingHooks(hooks ...CheckpointingHooks) MultiCheckpointingHooks {
	return hooks
}

func (h MultiCheckpointingHooks) AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error {
	for i := range h {
		if err := h[i].AfterBlsKeyRegistered(ctx, valAddr); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointConfirmed(ctx, epoch); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointForgotten(ctx sdk.Context, ckpt *RawCheckpoint) error {
	for i := range h {
		return h[i].AfterRawCheckpointForgotten(ctx, ckpt)
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointFinalized(ctx sdk.Context, epoch uint64) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointFinalized(ctx, epoch); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointBlsSigVerified(ctx sdk.Context, ckpt *RawCheckpoint) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointBlsSigVerified(ctx, ckpt); err != nil {
			return err
		}
	}
	return nil
}
