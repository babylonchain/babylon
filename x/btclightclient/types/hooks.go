package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ BTCLightClientHooks = &MultiBTCLightClientHooks{}

type MultiBTCLightClientHooks []BTCLightClientHooks

func NewMultiBTCLightClientHooks(hooks ...BTCLightClientHooks) MultiBTCLightClientHooks {
	return hooks
}

func (h MultiBTCLightClientHooks) AfterBTCHeaderInserted(ctx sdk.Context, headerInfo *BTCHeaderInfo) {
	for i := range h {
		h[i].AfterBTCHeaderInserted(ctx, headerInfo)
	}
}

func (h MultiBTCLightClientHooks) AfterBTCRollBack(ctx sdk.Context, headerInfo *BTCHeaderInfo) {
	for i := range h {
		h[i].AfterBTCRollBack(ctx, headerInfo)
	}
}

func (h MultiBTCLightClientHooks) AfterBTCRollForward(ctx sdk.Context, headerInfo *BTCHeaderInfo) {
	for i := range h {
		h[i].AfterBTCRollForward(ctx, headerInfo)
	}
}
