package keeper

import (
	etypes "github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Helper interface to be sure Hooks implement both epoching and light client hooks
type HandledHooks interface {
	etypes.EpochingHooks
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
