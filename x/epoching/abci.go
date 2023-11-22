package epoching

import (
	"context"
	"fmt"
	"time"

	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock,
// - record the current AppHash
// - if reaching the epoch beginning, then
//   - increment epoch number
//   - trigger AfterEpochBegins hook
//   - emit BeginEpoch event
//
// - if reaching the sealer header, i.e., the 2nd header of a non-zero epoch, then
//   - record the sealer header for the previous epoch
//
// NOTE: we follow Cosmos SDK's slashing/evidence modules for MVP. No need to modify them at the moment.
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// record the current AppHash
	k.RecordAppHash(ctx)

	// if this block is the first block of the next epoch
	// note that we haven't incremented the epoch number yet
	epoch := k.GetEpoch(ctx)
	if epoch.IsFirstBlockOfNextEpoch(ctx) {
		// increase epoch number
		incEpoch := k.IncEpoch(ctx)
		// init the msg queue of this new epoch
		k.InitMsgQueue(ctx)
		// init the slashed voting power of this new epoch
		k.InitSlashedVotingPower(ctx)
		// store the current validator set
		k.InitValidatorSet(ctx)
		// trigger AfterEpochBegins hook
		k.AfterEpochBegins(ctx, incEpoch.EpochNumber)
		// emit BeginEpoch event
		err := sdkCtx.EventManager().EmitTypedEvent(
			&types.EventBeginEpoch{
				EpochNumber: incEpoch.EpochNumber,
			},
		)
		if err != nil {
			return err
		}
	}

	if epoch.IsSecondBlock(ctx) {
		k.RecordSealerHeaderForPrevEpoch(ctx)
	}
	return nil
}

// EndBlocker is called at the end of every block.
// If reaching an epoch boundary, then
// - forward validator-related msgs (bonded -> unbonding) to the staking module
// - trigger AfterEpochEnds hook
// - emit EndEpoch event
// NOTE: The epoching module is not responsible for checkpoint-assisted unbonding (unbonding -> unbonded). Instead, it wraps the staking module and exposes interfaces to the checkpointing module. The checkpointing module will do the actual checkpoint-assisted unbonding upon each EndBlock.
func EndBlocker(ctx context.Context, k keeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validatorSetUpdate := []abci.ValidatorUpdate{}

	// if reaching an epoch boundary, then
	epoch := k.GetEpoch(ctx)
	if epoch.IsLastBlock(ctx) {
		// finalise this epoch, i.e., record the current header and the Merkle root of all AppHashs in this epoch
		if err := k.RecordLastHeaderAndAppHashRoot(ctx); err != nil {
			return nil, err
		}
		// get all msgs in the msg queue
		queuedMsgs := k.GetCurrentEpochMsgs(ctx)
		// forward each msg in the msg queue to the right keeper
		for _, msg := range queuedMsgs {
			res, err := k.HandleQueuedMsg(ctx, msg)
			// skip this failed msg and emit and event signalling it
			// we do not panic here as some users may wrap an invalid message
			// (e.g., self-delegate coins more than its balance, wrong coding of addresses, ...)
			// honest validators will have consistent execution results on the queued messages
			if err != nil {
				// emit an event signalling the failed execution
				err := sdkCtx.EventManager().EmitTypedEvent(
					&types.EventHandleQueuedMsg{
						EpochNumber: epoch.EpochNumber,
						Height:      msg.BlockHeight,
						TxId:        msg.TxId,
						MsgId:       msg.MsgId,
						Error:       err.Error(),
					},
				)
				if err != nil {
					return nil, err
				}
				// skip this failed msg
				continue
			}
			// for each event, emit an wrapped event EventTypeHandleQueuedMsg, which attaches the original attributes plus the original event type, the epoch number, txid and msgid to the event here
			for _, event := range res.Events {
				err := sdkCtx.EventManager().EmitTypedEvent(
					&types.EventHandleQueuedMsg{
						OriginalEventType:  event.Type,
						EpochNumber:        epoch.EpochNumber,
						TxId:               msg.TxId,
						MsgId:              msg.MsgId,
						OriginalAttributes: event.Attributes,
					},
				)
				if err != nil {
					return nil, err
				}
			}
		}

		// update validator set
		validatorSetUpdate = k.ApplyAndReturnValidatorSetUpdates(ctx)
		sdkCtx.Logger().Info(fmt.Sprintf("Epoching: validator set update of epoch %d: %v", epoch.EpochNumber, validatorSetUpdate))

		// trigger AfterEpochEnds hook
		k.AfterEpochEnds(ctx, epoch.EpochNumber)
		// emit EndEpoch event
		err := sdkCtx.EventManager().EmitTypedEvent(
			&types.EventEndEpoch{
				EpochNumber: epoch.EpochNumber,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return validatorSetUpdate, nil
}
