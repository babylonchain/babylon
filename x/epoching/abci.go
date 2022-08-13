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

	// if this block is the first block of the next epoch
	// note that we haven't incremented the epoch number yet
	epoch := k.GetEpoch(ctx)
	if epoch.IsFirstBlockOfNextEpoch(ctx) {
		// increase epoch number
		IncEpoch := k.IncEpoch(ctx)
		// init the slashed voting power of this new epoch
		k.InitSlashedVotingPower(ctx)
		// store the current validator set
		k.InitValidatorSet(ctx)
		// trigger AfterEpochBegins hook
		k.AfterEpochBegins(ctx, IncEpoch.EpochNumber)
		// emit BeginEpoch event
		err := ctx.EventManager().EmitTypedEvent(
			&types.EventBeginEpoch{
				EpochNumber: IncEpoch.EpochNumber,
			},
		)
		if err != nil {
			panic(err)
		}
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

	validatorSetUpdate := []abci.ValidatorUpdate{}

	// if reaching an epoch boundary, then
	epoch := k.GetEpoch(ctx)
	if epoch.IsLastBlock(ctx) {
		k.HandleQueuedMsgs(ctx, epoch)
		validatorSetUpdate = k.ApplyAndReturnValidatorSetUpdates(ctx)
	}

	return validatorSetUpdate
}
