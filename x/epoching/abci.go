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

func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	logger := k.Logger(ctx)
	logger.Info("unimplemented")

	// TODO: unimplemented:
	// - increment epoch number
	// - slashing equivocating/unlive validators following evidence/slashing modules
	// - trigger hooks and emit events
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	defer telemetry.ModuleMeasureSince(stakingtypes.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx)

	epochBoundary, err := k.GetEpochBoundary(ctx)
	if err != nil {
		logger.Error("failed to execute epochBoundary", err)
		return nil
	}

	// if reaching an epoch boundary, then
	// - forward validator-related msgs (bonded -> unbonding) to the staking module
	// - trigger AfterEpochEnds hook
	// - emit EndEpoch event
	if uint64(ctx.BlockHeight()) == epochBoundary.Uint64() {
		// get all msgs in the msg queue
		queuedMsgs, err := k.GetEpochMsgs(ctx)
		if err != nil {
			logger.Error("failed to execute GetEpochMsgs", err)
			return nil
		}
		// forward each msg in the msg queue to the right keeper
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
		// cleanup the current msg queue
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
		if k.AfterEpochEnds(ctx, epochNumber); err != nil {
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

	// - TODO: if an epoch is newly checkpointed, make unbonding validators/delegations in this epoch unbonded
	return nil
}
