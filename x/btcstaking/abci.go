package btcstaking

import (
	"context"
	"time"

	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	return k.BeginBlocker(ctx)
}

func EndBlocker(ctx context.Context, k keeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return []abci.ValidatorUpdate{}, nil
}
