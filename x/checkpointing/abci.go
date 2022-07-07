package checkpointing

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"time"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the second block after the epoch begins, then
// - extract the LastCommitHash from the block
// - create a raw checkpoint with the status of ACCUMULATING
// - start a BLS signer which creates a BLS sig transaction and distributes it to the network

func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// get the height of the last block in this epoch
	epochBoundary := k.GetEpochBoundary(ctx)
	// if this block is the second block of an epoch
	if uint64(ctx.BlockHeight())-2 == epochBoundary.Uint64() {
		// note that this epochNum is obtained before the BeginBlocker of the epoching module is executed
		// meaning that the epochNum has not been incremented upon a new epoch
		epochNum := k.GetEpochNumber(ctx)
		lch := ctx.BlockHeader().LastCommitHash
		err := k.BuildRawCheckpoint(ctx, epochNum, lch)
		if err != nil {
			panic("failed to generate a raw checkpoint")
		}

		// emit BeginEpoch event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRawCheckpointGenerated,
				sdk.NewAttribute(types.AttributeKeyEpochNumber, k.GetEpochNumber(ctx).String()),
			),
		})

		// TODO: call BLS signer to send a BLS-sig transaction
	}
}
