package keeper

import (
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// ApplyMatureUnbonding
// - unbonds all mature validators/delegations, and
// - finishes all mature redelegations
// in the corresponding queues, where
// - an unbonding/redelegation becomes mature when its corresponding epoch and all previous epochs have been checkpointed.
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/staking/keeper/val_state_change.go#L32-L91)
func ApplyMatureUnbonding(ctx sdk.Context, stk *stakingkeeper.Keeper, epochBoundaryHeader tmproto.Header) {
	currHeader := ctx.BlockHeader()

	// unbond all mature validators till the epoch boundary from the unbonding queue
	ctx.WithBlockHeader(epochBoundaryHeader)
	stk.UnbondAllMatureValidators(ctx)
	ctx.WithBlockHeader(currHeader)

	// get all mature unbonding delegations the epoch boundary from the ubd queue.
	ctx.WithBlockHeader(epochBoundaryHeader)
	matureUnbonds := stk.DequeueAllMatureUBDQueue(ctx, epochBoundaryHeader.Time)
	ctx.WithBlockHeader(currHeader)
	// unbond all mature delegations
	for _, dvPair := range matureUnbonds {
		addr, err := sdk.ValAddressFromBech32(dvPair.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delegatorAddress, err := sdk.AccAddressFromBech32(dvPair.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		balances, err := stk.CompleteUnbonding(ctx, delegatorAddress, addr)
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

	// get all mature redelegations till the epoch boundary from the red queue.
	ctx.WithBlockHeader(epochBoundaryHeader)
	matureRedelegations := stk.DequeueAllMatureRedelegationQueue(ctx, epochBoundaryHeader.Time)
	ctx.WithBlockHeader(currHeader)
	// finish all mature redelegations
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
		balances, err := stk.CompleteRedelegation(
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
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/staking/keeper/val_state_change.go#L18-L30)
func ApplyAndReturnValidatorSetUpdates(ctx sdk.Context, stk *stakingkeeper.Keeper) []abci.ValidatorUpdate {
	validatorUpdates, err := stk.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}

	return validatorUpdates
}
