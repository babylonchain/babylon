package keeper

import (
	"time"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ApplyMatureUnbonding
// - unbonds all mature validators/delegations, and
// - finishes all mature redelegations
// in the corresponding queues, where
// - an unbonding/redelegation becomes mature when its corresponding epoch and all previous epochs have been checkpointed.
func ApplyMatureUnbonding(ctx sdk.Context, sk *stakingkeeper.Keeper, epochBoundaryTime time.Time) {
	// unbond all mature validators from the unbonding queue
	sk.UnbondAllMatureValidators(ctx)

	// Remove all mature unbonding delegations from the ubd queue.
	// TODO: DequeueAllMatureUBDQueue does not make use of `currTime` parameter. Double-check
	matureUnbonds := sk.DequeueAllMatureUBDQueue(ctx, epochBoundaryTime)
	for _, dvPair := range matureUnbonds {
		addr, err := sdk.ValAddressFromBech32(dvPair.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delegatorAddress, err := sdk.AccAddressFromBech32(dvPair.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		balances, err := sk.CompleteUnbonding(ctx, delegatorAddress, addr)
		if err != nil {
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddress),
			),
		)
	}

	// Remove all mature redelegations from the red queue.
	// TODO: DequeueAllMatureRedelegationQueue does not make use of `currTime` parameter. Double-check
	matureRedelegations := sk.DequeueAllMatureRedelegationQueue(ctx, epochBoundaryTime)
	for _, dvvTriplet := range matureRedelegations {
		valSrcAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorSrcAddress)
		if err != nil {
			panic(err)
		}
		valDstAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorDstAddress)
		if err != nil {
			panic(err)
		}
		delegatorAddress, err := sdk.AccAddressFromBech32(dvvTriplet.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		balances, err := sk.CompleteRedelegation(
			ctx,
			delegatorAddress,
			valSrcAddr,
			valDstAddr,
		)
		if err != nil {
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddress),
				sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddress),
			),
		)
	}
}

// ApplyAndReturnValidatorSetUpdates applies and return accumulated updates to the bonded validator set, including
// * Updates the active validator set as keyed by LastValidatorPowerKey.
// * Updates the total power as keyed by LastTotalPowerKey.
// * Updates validator status' according to updated powers.
// * Updates the fee pool bonded vs not-bonded tokens.
// * Updates relevant indices.
// Triggered upon every epoch.
func ApplyAndReturnValidatorSetUpdates(ctx sdk.Context, sk *stakingkeeper.Keeper) []abci.ValidatorUpdate {
	validatorUpdates, err := sk.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}

	return validatorUpdates
}
