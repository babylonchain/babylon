package types

import (
	"context"
)

var _ BTCLightClientHooks = &MultiBTCLightClientHooks{}

type MultiBTCLightClientHooks []BTCLightClientHooks

func NewMultiBTCLightClientHooks(hooks ...BTCLightClientHooks) MultiBTCLightClientHooks {
	return hooks
}

func (h MultiBTCLightClientHooks) AfterBTCHeaderInserted(ctx context.Context, headerInfo *BTCHeaderInfo) {
	for i := range h {
		h[i].AfterBTCHeaderInserted(ctx, headerInfo)
	}
}

func (h MultiBTCLightClientHooks) AfterBTCRollBack(ctx context.Context, headerInfo *BTCHeaderInfo) {
	for i := range h {
		h[i].AfterBTCRollBack(ctx, headerInfo)
	}
}

func (h MultiBTCLightClientHooks) AfterBTCRollForward(ctx context.Context, headerInfo *BTCHeaderInfo) {
	for i := range h {
		h[i].AfterBTCRollForward(ctx, headerInfo)
	}
}
