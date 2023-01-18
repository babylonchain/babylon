package keeper

import (
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	etypes "github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Helper interface to be sure Hooks implement both epoching and light client hooks
type HandledHooks interface {
	etypes.EpochingHooks
	checkpointingtypes.CheckpointingHooks
}

type Hooks struct {
	k Keeper
}

var _ HandledHooks = Hooks{}

func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterEpochBegins(ctx sdk.Context, epoch uint64) {}

func (h Hooks) AfterEpochEnds(ctx sdk.Context, epoch uint64) {
	h.k.updateBtcLightClientHeightForEpoch(ctx, epoch)
}

func (h Hooks) BeforeSlashThreshold(ctx sdk.Context, valSet etypes.ValidatorSet) {}

func (h Hooks) AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error {
	return nil
}
func (h Hooks) AfterRawCheckpointFinalized(ctx sdk.Context, epoch uint64) error {
	return nil
}

func (h Hooks) AfterRawCheckpointBlsSigVerified(ctx sdk.Context, ckpt *checkpointingtypes.RawCheckpoint) {
	h.k.updateBtcLightClientHeightForCheckpoint(ctx, ckpt)
}
