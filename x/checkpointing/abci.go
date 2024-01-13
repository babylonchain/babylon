package checkpointing

import (
	"context"
	"fmt"
	"time"

	"github.com/babylonchain/babylon/x/checkpointing/types"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the first block after the epoch begins
// then we store the current validator set with BLS keys
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	epoch := k.GetEpoch(ctx)
	if epoch.IsFirstBlock(ctx) {
		err := k.InitValidatorBLSSet(ctx)
		if err != nil {
			panic(fmt.Errorf("failed to store validator BLS set: %w", err))
		}
	}
	return nil
}
