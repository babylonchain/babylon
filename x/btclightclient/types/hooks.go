package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ BTCLightClientHooks = &MultiBTCLightClientHooks{}

type MultiBTCLightClientHooks []BTCLightClientHooks

func NewMultiBTCLightClientHooks(hooks ...BTCLightClientHooks) MultiBTCLightClientHooks {
	return hooks
}

func (h MultiBTCLightClientHooks) AfterTipUpdated(ctx sdk.Context, height uint64) {
	for i := range h {
		h[i].AfterTipUpdated(ctx, height)
	}
}
