package keeper

import (
	"context"
	"fmt"

	bbntypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
)

var _ types.FinalityHooks = Hooks{}

// Hooks wrapper struct for BTC staking keeper
type Hooks struct {
	k Keeper
}

// Return the finality hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterInactiveFinalityProviderDetected updates the status of the given finality provider to `inactive`
func (h Hooks) AfterInactiveFinalityProviderDetected(ctx context.Context, fpPk *bbntypes.BIP340PubKey) error {
	fp, err := h.k.GetFinalityProvider(ctx, fpPk.MustMarshal())
	if err != nil {
		return err
	}

	if fp.IsInactive() {
		return fmt.Errorf("the finality provider %s is already detected as inactive", fpPk.MarshalHex())
	}

	fp.Inactive = true

	h.k.SetFinalityProvider(ctx, fp)

	return nil
}
