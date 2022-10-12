package checkpointing

import (
	"fmt"
	"github.com/babylonchain/babylon/types/retry"
	"time"

	"github.com/babylonchain/babylon/x/checkpointing/types"

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

	// if this block is the second block of an epoch
	epoch := k.GetEpoch(ctx)
	if epoch.IsSecondBlock(ctx) {
		// note that this epochNum is obtained after the BeginBlocker of the epoching module is executed
		// meaning that the epochNum has been incremented upon a new epoch
		lch := ctx.BlockHeader().LastCommitHash
		ckpt, err := k.BuildRawCheckpoint(ctx, epoch.EpochNumber-1, lch)
		if err != nil {
			panic("failed to generate a raw checkpoint")
		}

		// emit BeginEpoch event
		err = ctx.EventManager().EmitTypedEvent(
			&types.EventCheckpointAccumulating{
				Checkpoint: ckpt,
			},
		)
		if err != nil {
			panic(err)
		}

		go func() {
			// TODO should read the parameters from config file
			err = retry.Do(1*time.Second, 1*time.Minute, func() error {
				_, err := k.SendBlsSig(ctx, epoch.EpochNumber-1, lch)
				if err != nil {
					return err
				}
				ctx.Logger().Info(fmt.Sprintf("Successfully sent BLS-sig tx for epoch %v", epoch.EpochNumber-1))
				return nil
			})
			if err != nil {
				// failing to send a BLS-sig causes a panicking situation
				panic(err)
			}
		}()
	}
}
