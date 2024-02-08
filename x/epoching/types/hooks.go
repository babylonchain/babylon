package types

import (
	"context"
)

// combine multiple Epoching hooks, all hook functions are run in array sequence
var _ EpochingHooks = &MultiEpochingHooks{}

type MultiEpochingHooks []EpochingHooks

func NewMultiEpochingHooks(hooks ...EpochingHooks) MultiEpochingHooks {
	return hooks
}

func (h MultiEpochingHooks) AfterEpochBegins(ctx context.Context, epoch uint64) {
	for i := range h {
		h[i].AfterEpochBegins(ctx, epoch)
	}
}

func (h MultiEpochingHooks) AfterEpochEnds(ctx context.Context, epoch uint64) {
	for i := range h {
		h[i].AfterEpochEnds(ctx, epoch)
	}
}

func (h MultiEpochingHooks) BeforeSlashThreshold(ctx context.Context, valSet ValidatorSet) {
	for i := range h {
		h[i].BeforeSlashThreshold(ctx, valSet)
	}
}
