package finality

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
)

func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// if the BTC staking protocol is activated, i.e., there exists a height where a finality provider
	// has voting power, start indexing and tallying blocks
	if _, err := k.BTCStakingKeeper.GetBTCStakingActivatedHeight(ctx); err == nil {
		// index the current block
		k.IndexBlock(ctx)
		// tally all non-finalised blocks
		k.TallyBlocks(ctx)
		// jail inactive finality providers if there are any
		// TODO: decide which height to use the handle liveness
		height := sdk.UnwrapSDKContext(ctx).HeaderInfo().Height
		k.HandleLiveness(ctx, height)
	}

	return []abci.ValidatorUpdate{}, nil
}
