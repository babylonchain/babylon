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
		// get epoch number
		epochNumber := k.GetEpochNumber(ctx)
		// get all msgs in the msg queue
		queuedMsgs := k.GetEpochMsgs(ctx)
		// forward each msg in the msg queue to the right keeper
		for _, msg := range queuedMsgs {
			res, err := k.HandleQueuedMsg(ctx, msg)
			// we should skip msg with errors rather than panicking, as some users may wrap an invalid message
			// (e.g., self-delegate coins more than its balance, wrong coding of addresses, ...)
			// honest validators will have consistent execution results on the queued messages
			if err != nil {
				logger.Error(err.Error())
				continue
			}
			// append the epoch info to each event and emit event
			for _, event := range res.Events {
				newAttr := sdk.NewAttribute(types.AttributeKeyEpoch, epochNumber.String()).ToKVPair()
				event.Attributes = append(event.Attributes, newAttr)
				typedEvent, err := sdk.ParseTypedEvent(event)
				if err != nil {
					logger.Error(err.Error())
					continue
				}
				if err := ctx.EventManager().EmitTypedEvent(typedEvent); err != nil {
					logger.Error(err.Error())
					continue
				}
			}
			logger.Info(res.Log)
		}

		// update validator set
		validatorSetUpdate = k.ApplyAndReturnValidatorSetUpdates(ctx)
		// clear the current msg queue
		k.ClearEpochMsgs(ctx)
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
