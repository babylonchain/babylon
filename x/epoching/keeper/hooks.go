package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

// Implements StakingHooks/EpochingHooks interfaces
var _ stakingtypes.StakingHooks = Hooks{}
var _ types.EpochingHooks = Keeper{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// AfterEpochBegins - call hook if registered
func (k Keeper) AfterEpochBegins(ctx sdk.Context, epoch sdk.Uint) {
	if k.hooks != nil {
		k.hooks.AfterEpochBegins(ctx, epoch)
	}
}

// AfterEpochEnds - call hook if registered
func (k Keeper) AfterEpochEnds(ctx sdk.Context, epoch sdk.Uint) {
	if k.hooks != nil {
		k.hooks.AfterEpochEnds(ctx, epoch)
	}
}

// BeforeSlashThreshold triggers the BeforeSlashThreshold hook for other modules that register this hook
func (k Keeper) BeforeSlashThreshold(ctx sdk.Context, valAddrs []sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.BeforeSlashThreshold(ctx, valAddrs)
	}
}

// BeforeValidatorSlashed records the slash event
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	logger := h.k.Logger(ctx)

	// add the validator address to the set
	if err := h.k.AddSlashedValidator(ctx, valAddr); err != nil {
		logger.Error("failed to execute AddSlashedValidator", err)
	}

	numSlashedVals, err := h.k.GetSlashedValidatorSetSize(ctx)
	if err != nil {
		logger.Error("failed to execute GetSlashedValidatorSetSize", err)
	}
	numMaxVals := h.k.stk.GetParams(ctx).MaxValidators
	// if a certain threshold (1/3 or 2/3) validators are slashed in a single epoch, emit event and trigger hook
	if numSlashedVals.Uint64() == uint64(numMaxVals)/3 || numSlashedVals.Uint64() == uint64(numMaxVals)*2/3 {
		// emit event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeSlashThreshold,
				sdk.NewAttribute(types.AttributeKeyNumSlashedVals, numSlashedVals.String()),
				sdk.NewAttribute(types.AttributeKeyNumMaxVals, fmt.Sprint(numMaxVals)),
			),
		})
		// trigger hook
		slashedVals, err := h.k.GetSlashedValidators(ctx)
		if err != nil {
			logger.Error("failed to execute GetSlashedValidators", err)
		}
		h.k.BeforeSlashThreshold(ctx, slashedVals)
	}
}

// Other staking hooks that are not used in the epoching module
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)   {}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
