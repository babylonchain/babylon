package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple Epoching hooks, all hook functions are run in array sequence
var _ EpochingHooks = &MultiEpochingHooks{}

type MultiEpochingHooks []EpochingHooks

func NewMultiEpochingHooks(hooks ...EpochingHooks) MultiEpochingHooks {
	return hooks
}

func (h MultiEpochingHooks) AfterEpochBegins(ctx sdk.Context, epoch sdk.Uint) {
	for i := range h {
		h[i].AfterEpochBegins(ctx, epoch)
	}
}

func (h MultiEpochingHooks) AfterEpochEnds(ctx sdk.Context, epoch sdk.Uint) {
	for i := range h {
		h[i].AfterEpochEnds(ctx, epoch)
	}
}

func (h MultiEpochingHooks) BeforeSlashThreshold(ctx sdk.Context, valAddrs []sdk.ValAddress) {
	for i := range h {
		h[i].BeforeSlashThreshold(ctx, valAddrs)
	}
}
