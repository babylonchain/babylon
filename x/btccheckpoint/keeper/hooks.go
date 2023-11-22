package keeper

import (
	"context"
	ltypes "github.com/babylonchain/babylon/x/btclightclient/types"
	etypes "github.com/babylonchain/babylon/x/epoching/types"
)

// HandledHooks Helper interface to ensure Hooks implements
// both epoching and btclightclient hooks
type HandledHooks interface {
	ltypes.BTCLightClientHooks
	etypes.EpochingHooks
}

type Hooks struct {
	k Keeper
}

var _ HandledHooks = Hooks{}

func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterBTCRollBack(ctx context.Context, _ *ltypes.BTCHeaderInfo) {
	h.k.setBtcLightClientUpdated(ctx)
}

func (h Hooks) AfterBTCRollForward(ctx context.Context, _ *ltypes.BTCHeaderInfo) {
	h.k.setBtcLightClientUpdated(ctx)
}

func (h Hooks) AfterBTCHeaderInserted(_ context.Context, _ *ltypes.BTCHeaderInfo) {}

func (h Hooks) AfterEpochBegins(_ context.Context, _ uint64) {}

func (h Hooks) AfterEpochEnds(_ context.Context, _ uint64) {}

func (h Hooks) BeforeSlashThreshold(_ context.Context, _ etypes.ValidatorSet) {}
