package finality

import (
	"time"

	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// if the BTC staking protocol is activated, i.e., there exists a height where a BTC validator
	// has voting power, start indexing and tallying blocks
	if _, err := k.BTCStakingKeeper.GetBTCStakingActivatedHeight(ctx); err == nil {
		// index the current block
		k.IndexBlock(ctx)
		// tally all non-finalised blocks
		k.TallyBlocks(ctx)
	}

	return []abci.ValidatorUpdate{}
}
