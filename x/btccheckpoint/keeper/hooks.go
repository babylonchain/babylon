package keeper

import (
	ltypes "github.com/babylonchain/babylon/x/btclightclient/types"
	etypes "github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Helper interface to be sure Hooks implement both epoching and light client hooks
type HandledHooks interface {
	ltypes.BTCLightClientHooks
	etypes.EpochingHooks
}

type Hooks struct {
	k Keeper
}

var _ HandledHooks = Hooks{}

func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterBTCRollBack(ctx sdk.Context, headerInfo *ltypes.BTCHeaderInfo) {
	h.k.setBtcLightClientUpdated(ctx)
}

func (h Hooks) AfterBTCRollForward(ctx sdk.Context, headerInfo *ltypes.BTCHeaderInfo) {
	h.k.setBtcLightClientUpdated(ctx)
}

func (h Hooks) AfterBTCHeaderInserted(ctx sdk.Context, headerInfo *ltypes.BTCHeaderInfo) {}

func (h Hooks) AfterEpochBegins(ctx sdk.Context, epoch uint64) {}

func (h Hooks) AfterEpochEnds(ctx sdk.Context, epoch uint64) {}

func (h Hooks) BeforeSlashThreshold(ctx sdk.Context, valSet etypes.ValidatorSet) {}
