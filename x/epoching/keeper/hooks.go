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
	thresholds := []float64{1 / 3, 2 / 3}

	epochNumber := h.k.GetEpochNumber(ctx)
	totalVotingPower := h.k.GetTotalVotingPower(ctx, epochNumber)
	validatorSet := h.k.GetValidatorSet(ctx, epochNumber)

	// calculate total slashed voting power
	slashedVotingPower := int64(0)
	slashedVals := h.k.GetSlashedValidators(ctx, epochNumber)
	for _, slashedVal := range slashedVals {
		if power, ok := validatorSet[slashedVal.String()]; ok {
			slashedVotingPower += power
		} else {
			panic(types.ErrUnknownValidator)
		}
	}

	// voting power of this validator
	thisVotingPower, ok := validatorSet[valAddr.String()]
	if !ok {
		panic(types.ErrUnknownValidator)
	}

	for _, threshold := range thresholds {
		// if a certain threshold voting power is slashed in a single epoch, emit event and trigger hook
		if float64(slashedVotingPower) < float64(totalVotingPower)*threshold && float64(totalVotingPower)*threshold <= float64(slashedVotingPower+thisVotingPower) {
			// add the newly slashed validator
			slashedVals = append(slashedVals, valAddr)
			// emit event
			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeSlashThreshold,
					sdk.NewAttribute(types.AttributeKeySlashedVotingPower, fmt.Sprintf("%d", slashedVotingPower)),
					sdk.NewAttribute(types.AttributeKeyTotalVotingPower, fmt.Sprintf("%d", slashedVotingPower)),
					sdk.NewAttribute(types.AttributeKeySlashedValidators, fmt.Sprintf("%v", slashedVals)),
				),
			})
			// trigger hook
			slashedVals := h.k.GetSlashedValidators(ctx, epochNumber)
			h.k.BeforeSlashThreshold(ctx, slashedVals)
		}
	}

	// add the validator address to the set
	h.k.AddSlashedValidator(ctx, valAddr)
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
