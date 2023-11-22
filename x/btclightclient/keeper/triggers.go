package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) triggerHeaderInserted(ctx context.Context, headerInfo *types.BTCHeaderInfo) {
	// Trigger AfterBTCHeaderInserted hook
	k.AfterBTCHeaderInserted(ctx, headerInfo)
	// Emit HeaderInserted event
	sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&types.EventBTCHeaderInserted{Header: headerInfo}) //nolint:errcheck
}

func (k Keeper) triggerRollBack(ctx context.Context, headerInfo *types.BTCHeaderInfo) {
	// Trigger AfterBTCRollBack hook
	k.AfterBTCRollBack(ctx, headerInfo)
	// Emit BTCRollBack event
	sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&types.EventBTCRollBack{Header: headerInfo}) //nolint:errcheck
}

func (k Keeper) triggerRollForward(ctx context.Context, headerInfo *types.BTCHeaderInfo) {
	// Trigger AfterBTCRollForward hook
	k.AfterBTCRollForward(ctx, headerInfo)
	// Emit BTCRollForward event
	sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&types.EventBTCRollForward{Header: headerInfo}) //nolint:errcheck
}
