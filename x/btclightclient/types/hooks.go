package types

import (
	bbl "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ BTCLightClientHooks = &MultiBTCLightClientHooks{}

type MultiBTCLightClientHooks []BTCLightClientHooks

func NewMultiBTCLightClientHooks(hooks ...BTCLightClientHooks) MultiBTCLightClientHooks {
	return hooks
}

func (h MultiBTCLightClientHooks) AfterBTCRollBack(ctx sdk.Context, hash bbl.BTCHeaderHashBytes, height uint64) {
	for i := range h {
		h[i].AfterBTCRollBack(ctx, hash, height)
	}
}

func (h MultiBTCLightClientHooks) AfterBTCRollForward(ctx sdk.Context, hash bbl.BTCHeaderHashBytes, height uint64) {
	for i := range h {
		h[i].AfterBTCRollForward(ctx, hash, height)
	}
}
