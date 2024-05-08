package keeper

import (
	"context"

	"cosmossdk.io/math"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ensures Keeper implements EpochingHooks interfaces
var _ types.EpochingHooks = Keeper{}

// AfterEpochBegins - call hook if registered
func (k Keeper) AfterEpochBegins(ctx context.Context, epoch uint64) {
	if k.hooks != nil {
		k.hooks.AfterEpochBegins(ctx, epoch)
	}
}

// AfterEpochEnds - call hook if registered
func (k Keeper) AfterEpochEnds(ctx context.Context, epoch uint64) {
	if k.hooks != nil {
		k.hooks.AfterEpochEnds(ctx, epoch)
	}
}

// BeforeSlashThreshold triggers the BeforeSlashThreshold hook for other modules that register this hook
func (k Keeper) BeforeSlashThreshold(ctx context.Context, valSet types.ValidatorSet) {
	if k.hooks != nil {
		k.hooks.BeforeSlashThreshold(ctx, valSet)
	}
}

// Wrapper struct
type Hooks struct {
	k Keeper
}

// ensures Hooks implements StakingHooks and CheckpointingHooks interfaces
var _ stakingtypes.StakingHooks = Hooks{}
var _ checkpointingtypes.CheckpointingHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// BeforeValidatorSlashed records the slash event
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	thresholds := []float64{float64(1) / float64(3), float64(2) / float64(3)}

	epochNumber := h.k.GetEpoch(ctx).EpochNumber
	totalVotingPower := h.k.GetTotalVotingPower(ctx, epochNumber)

	// calculate total slashed voting power
	slashedVotingPower := h.k.GetSlashedVotingPower(ctx, epochNumber)
	// voting power of this validator
	thisVotingPower, err := h.k.GetValidatorVotingPower(ctx, epochNumber, valAddr)
	thisVal := types.Validator{Addr: valAddr, Power: thisVotingPower}
	if err != nil {
		// It's possible that the most powerful validator outside the validator set enrols to the validator after this validator is slashed.
		// Consequently, here we cannot find this validator in the validatorSet map.
		// As we consider the validator set in the epoch beginning to be the validator set throughout this epoch, we consider this new validator in the edge to have no voting power and return directly here.
		return err
	}

	for _, threshold := range thresholds {
		// if a certain threshold voting power is slashed in a single epoch, emit event and trigger hook
		if float64(slashedVotingPower) < float64(totalVotingPower)*threshold && float64(totalVotingPower)*threshold <= float64(slashedVotingPower+thisVotingPower) {
			slashedVals := h.k.GetSlashedValidators(ctx, epochNumber)
			slashedVals = append(slashedVals, thisVal)
			event := types.NewEventSlashThreshold(slashedVotingPower, totalVotingPower, slashedVals)
			if err := sdkCtx.EventManager().EmitTypedEvent(&event); err != nil {
				panic(err)
			}
			h.k.BeforeSlashThreshold(ctx, slashedVals)
		}
	}

	// add the validator address to the set
	if err := h.k.AddSlashedValidator(ctx, valAddr); err != nil {
		// same as above error case
		return err
	}

	return nil
}

func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return h.k.RecordNewValState(sdk.UnwrapSDKContext(ctx), valAddr, types.BondState_CREATED)
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.RecordNewValState(sdk.UnwrapSDKContext(ctx), valAddr, types.BondState_REMOVED)
}

func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.RecordNewValState(sdk.UnwrapSDKContext(ctx), valAddr, types.BondState_BONDED)
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.RecordNewValState(sdk.UnwrapSDKContext(ctx), valAddr, types.BondState_UNBONDING)
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return h.k.RecordNewDelegationState(sdk.UnwrapSDKContext(ctx), delAddr, valAddr, nil, types.BondState_REMOVED)
}

func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	return nil
}

// Other hooks that are not used in the epoching module
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// Checkpointing hooks
func (h Hooks) AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error {
	// finalise all unbonding validators/delegations in this epoch
	h.k.ApplyMatureUnbonding(ctx, epoch)
	return nil
}

func (h Hooks) AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error { return nil }

func (h Hooks) AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error    { return nil }
func (h Hooks) AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error { return nil }
func (h Hooks) AfterRawCheckpointForgotten(ctx context.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return nil
}

func (h Hooks) AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return nil
}
