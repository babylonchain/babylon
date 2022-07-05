package epoching

import (
	"time"

	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the epoch beginning, then
//    - increment epoch number
//    - trigger AfterEpochBegins hook
//    - emit BeginEpoch event
// NOTE: we follow Cosmos SDK's slashing/evidence modules for MVP. No need to modify them at the moment.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// get the height of the last block in this epoch
	epochBoundary := k.GetEpochBoundary(ctx)
	// if this block is the first block of an epoch
	// note that we haven't incremented the epoch number yet
	if uint64(ctx.BlockHeight())-1 == epochBoundary.Uint64() {
		// increase epoch number
		incEpochNumber := k.IncEpochNumber(ctx)
		// trigger AfterEpochBegins hook
		k.AfterEpochBegins(ctx, incEpochNumber)
		// emit BeginEpoch event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeBeginEpoch,
				sdk.NewAttribute(types.AttributeKeyEpoch, incEpochNumber.String()),
			),
		})
	}
}

// EndBlocker is called at the end of every block.
// If reaching an epoch boundary, then
// - forward validator-related msgs (bonded -> unbonding) to the staking module
// - trigger AfterEpochEnds hook
// - emit EndEpoch event
// NOTE: The epoching module is not responsible for checkpoint-assisted unbonding (unbonding -> unbonded). Instead, it wraps the staking module and exposes interfaces to the checkpointing module. The checkpointing module will do the actual checkpoint-assisted unbonding upon each EndBlock.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx)
	validatorSetUpdate := []abci.ValidatorUpdate{}

	// get the height of the last block in this epoch
	epochBoundary := k.GetEpochBoundary(ctx)

	// if reaching an epoch boundary, then
	if uint64(ctx.BlockHeight()) == epochBoundary.Uint64() {
		// get all msgs in the msg queue
		queuedMsgs := k.GetEpochMsgs(ctx)
		// forward each msg in the msg queue to the right keeper
		// TODO: is it possible or beneficial if we can get the execution results of the delayed messages in the epoching module rather than in the staking module?
		for _, msg := range queuedMsgs {
			res := k.HandleQueuedMsg(ctx, msg)
			// TODO: what to do on the events and logs returned by the staking module?
			logger.Info(res.Log)
		}

		// update validator set
		validatorSetUpdate = k.ApplyAndReturnValidatorSetUpdates(ctx)
		// clear the current msg queue
		k.ClearEpochMsgs(ctx)
		// clear the slashed validator set
		k.ClearSlashedValidators(ctx)
		// get epoch number
		epochNumber := k.GetEpochNumber(ctx)
		// trigger AfterEpochEnds hook
		k.AfterEpochEnds(ctx, epochNumber)
		// emit EndEpoch event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeEndEpoch,
				sdk.NewAttribute(types.AttributeKeyEpoch, epochNumber.String()),
			),
		})
	}

	return validatorSetUpdate
}
