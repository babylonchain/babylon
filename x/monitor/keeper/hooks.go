package keeper

import (
	"context"

	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	etypes "github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HandledHooks Helper interface to be sure Hooks implement both epoching and light client hooks
type HandledHooks interface {
	etypes.EpochingHooks
	checkpointingtypes.CheckpointingHooks
}

type Hooks struct {
	k Keeper
}

// Hooks Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterEpochBegins(ctx context.Context, epoch uint64) {}

func (h Hooks) AfterEpochEnds(ctx context.Context, epoch uint64) {
	h.k.updateBtcLightClientHeightForEpoch(ctx, epoch)
}

func (h Hooks) BeforeSlashThreshold(ctx context.Context, valSet etypes.ValidatorSet) {}

func (h Hooks) AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error {
	return nil
}
func (h Hooks) AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error {
	return nil
}

func (h Hooks) AfterRawCheckpointForgotten(ctx context.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return h.k.removeCheckpointRecord(ctx, ckpt)
}

func (h Hooks) AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error {
	return nil
}

func (h Hooks) AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return h.k.updateBtcLightClientHeightForCheckpoint(ctx, ckpt)
}
