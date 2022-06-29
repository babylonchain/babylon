package keeper

import (
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

// AfterEpochBegins triggers the AfterEpochBegins hook for other modules that register this hook
func (k Keeper) AfterEpochBegins(ctx sdk.Context, epoch sdk.Uint) error {
	if k.hooks != nil {
		return k.hooks.AfterEpochBegins(ctx, epoch)
	}
	return nil
}

// AfterEpochEnds triggers the AfterEpochEnds hook for other modules that register this hook
func (k Keeper) AfterEpochEnds(ctx sdk.Context, epoch sdk.Uint) error {
	if k.hooks != nil {
		return k.hooks.AfterEpochEnds(ctx, epoch)
	}
	return nil
}

// BeforeValidatorSlashed records the slash event
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// TODO: unimplemented:
	//  - add the validator address to the set
	//  - if = 1/3 or 2/3 validators are slashed in a single epoch, emit event and trigger hook
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
