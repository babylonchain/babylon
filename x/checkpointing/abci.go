package checkpointing

import (
	"context"
	"fmt"
	"time"

	"github.com/babylonchain/babylon/x/checkpointing/types"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the second block after the epoch begins, then
// - extract the AppHash from the block
// - create a raw checkpoint with the status of ACCUMULATING
// - start a BLS signer which creates a BLS sig transaction and distributes it to the network
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// if this block is the second block of an epoch
	epoch := k.GetEpoch(ctx)
	if epoch.IsFirstBlock(ctx) {
		err := k.InitValidatorBLSSet(ctx)
		if err != nil {
			panic(fmt.Errorf("failed to store validator BLS set: %w", err))
		}
	}
	if epoch.IsSecondBlock(ctx) {
		// note that this epochNum is obtained after the BeginBlocker of the epoching module is executed
		// meaning that the epochNum has been incremented upon a new epoch

		appHash := sdkCtx.HeaderInfo().AppHash
		ckpt, err := k.BuildRawCheckpoint(ctx, epoch.EpochNumber-1, appHash)
		if err != nil {
			panic("failed to generate a raw checkpoint")
		}

		// emit BeginEpoch event
		err = sdkCtx.EventManager().EmitTypedEvent(
			&types.EventCheckpointAccumulating{
				Checkpoint: ckpt,
			},
		)
		if err != nil {
			panic(err)
		}
		curValSet := k.GetValidatorSet(ctx, epoch.EpochNumber-1)

		go func() {
			err := k.SendBlsSig(ctx, epoch.EpochNumber-1, appHash, curValSet)
			if err != nil {
				// failing to send a BLS-sig causes a panicking situation
				panic(err)
			}
		}()
	}
	return nil
}
