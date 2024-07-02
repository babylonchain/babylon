package types

import (
	"context"

	"github.com/babylonchain/babylon/types"
)

// combine multiple BTC staking hooks, all hook functions are run in array sequence
var _ BtcStakingHooks = &MultiBtcStakingHooks{}

type MultiBtcStakingHooks []BtcStakingHooks

func NewMultiBtcStakingHooks(hooks ...BtcStakingHooks) MultiBtcStakingHooks {
	return hooks
}

func (h MultiBtcStakingHooks) AfterFinalityProviderActivated(ctx context.Context, btcPk *types.BIP340PubKey) error {
	for i := range h {
		if err := h[i].AfterFinalityProviderActivated(ctx, btcPk); err != nil {
			return err
		}
	}

	return nil
}
