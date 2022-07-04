package epoching

import (
	"time"

	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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

	logger := k.Logger(ctx)

	// get the height of the last block in this epoch
	epochBoundary, err := k.GetEpochBoundary(ctx)
	if err != nil {
		logger.Error("failed to execute GetEpochBoundary", err)
	}
	// if this block is the first block of an epoch
	// note that we haven't incremented the epoch number yet
	if uint64(ctx.BlockHeight())-1 == epochBoundary.Uint64() {
		// increase epoch number
		incEpochNumber, err := k.IncEpochNumber(ctx)
		if err != nil {
			logger.Error("failed to execute IncEpochNumber", err)
		}
		// trigger AfterEpochBegins hook
		if err := k.AfterEpochBegins(ctx, incEpochNumber); err != nil {
			logger.Error("failed to execute AfterEpochBegins", err)
		}
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
	defer telemetry.ModuleMeasureSince(stakingtypes.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx)

	// get the height of the last block in this epoch
	epochBoundary, err := k.GetEpochBoundary(ctx)
	if err != nil {
		logger.Error("failed to execute GetEpochBoundary", err)
		return []abci.ValidatorUpdate{}
	}

	// if reaching an epoch boundary, then
	if uint64(ctx.BlockHeight()) == epochBoundary.Uint64() {
		// get all msgs in the msg queue
		queuedMsgs, err := k.GetEpochMsgs(ctx)
		if err != nil {
			logger.Error("failed to execute GetEpochMsgs", err)
			return nil
		}
		// forward each msg in the msg queue to the right keeper
		// TODO: is it possible or beneficial if we can get the execution results of the delayed messages in the epoching module rather than in the staking module?
		for _, msg := range queuedMsgs {
			switch unwrappedMsg := msg.Msg.(type) {
			case *types.QueuedMessage_MsgCreateValidator:
				unwrappedMsgWithType := unwrappedMsg.MsgCreateValidator
				if _, err := k.StakingMsgServer.CreateValidator(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
					logger.Error("failed to forward MsgCreateValidator", err)
				}
			case *types.QueuedMessage_MsgDelegate:
				unwrappedMsgWithType := unwrappedMsg.MsgDelegate
				if _, err := k.StakingMsgServer.Delegate(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
					logger.Error("failed to forward MsgDelegate", err)
				}
			case *types.QueuedMessage_MsgUndelegate:
				unwrappedMsgWithType := unwrappedMsg.MsgUndelegate
				if _, err := k.StakingMsgServer.Undelegate(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
					logger.Error("failed to forward MsgUndelegate", err)
				}
			case *types.QueuedMessage_MsgBeginRedelegate:
				unwrappedMsgWithType := unwrappedMsg.MsgBeginRedelegate
				if _, err := k.StakingMsgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
					logger.Error("failed to forward MsgBeginRedelegate", err)
				}
			default:
				logger.Error("unknown type of QueuedMessage: %v", msg)
				return nil
			}
		}
		// clear the current msg queue
		if err := k.ClearEpochMsgs(ctx); err != nil {
			logger.Error("failed to execute ClearEpochMsgs", err)
			return nil
		}

		// get epoch number
		epochNumber, err := k.GetEpochNumber(ctx)
		if err != nil {
			logger.Error("failed to execute GetEpochNumber", err)
			return nil
		}
		// trigger AfterEpochEnds hook
		if err := k.AfterEpochEnds(ctx, epochNumber); err != nil {
			logger.Error("failed to execute GetEpochNumber", err)
			return nil
		}
		// emit EndEpoch event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeEndEpoch,
				sdk.NewAttribute(types.AttributeKeyEpoch, epochNumber.String()),
			),
		})
	}

	return nil
}
