package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bbntypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
)

var _ types.BtcStakingHooks = Hooks{}

// Hooks wrapper struct for finality keeper
type Hooks struct {
	k Keeper
}

// Return the BTC staking hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterFinalityProviderActivated updates the signing info start height or create a new signing info
func (h Hooks) AfterFinalityProviderActivated(ctx context.Context, fpPk *bbntypes.BIP340PubKey) error {
	signingInfo, err := h.k.FinalityProviderSigningTracker.Get(ctx, fpPk.MustMarshal())
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err == nil {
		signingInfo.StartHeight = sdkCtx.BlockHeight()
	} else if errors.Is(err, collections.ErrNotFound) {
		signingInfo = types.NewFinalityProviderSigningInfo(
			fpPk,
			sdkCtx.BlockHeight(),
			0,
		)
	}

	return h.k.FinalityProviderSigningTracker.Set(ctx, fpPk.MustMarshal(), signingInfo)
}
