package incentive

import (
	"context"
	"time"

	"github.com/babylonchain/babylon/x/incentive/keeper"
	"github.com/babylonchain/babylon/x/incentive/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// handle coins in the fee collector account, including
	// - send a portion of coins in the fee collector account to the incentive module account
	// - accumulate BTC staking gauge at the current height
	// - accumulate BTC timestamping gauge at the current epoch
	if sdk.UnwrapSDKContext(ctx).HeaderInfo().Height > 0 {
		k.HandleCoinsInFeeCollector(ctx)
	}
	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return []abci.ValidatorUpdate{}, nil
}
