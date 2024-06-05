package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple Checkpointing hooks, all hook functions are run in array sequence
var _ CheckpointingHooks = &MultiCheckpointingHooks{}

type MultiCheckpointingHooks []CheckpointingHooks

func NewMultiCheckpointingHooks(hooks ...CheckpointingHooks) MultiCheckpointingHooks {
	return hooks
}

func (h MultiCheckpointingHooks) AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error {
	for i := range h {
		if err := h[i].AfterBlsKeyRegistered(ctx, valAddr); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointSealed(ctx, epoch); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointConfirmed(ctx, epoch); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointForgotten(ctx context.Context, ckpt *RawCheckpoint) error {
	for i := range h {
		return h[i].AfterRawCheckpointForgotten(ctx, ckpt)
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointFinalized(ctx, epoch); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiCheckpointingHooks) AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *RawCheckpoint) error {
	for i := range h {
		if err := h[i].AfterRawCheckpointBlsSigVerified(ctx, ckpt); err != nil {
			return err
		}
	}
	return nil
}
