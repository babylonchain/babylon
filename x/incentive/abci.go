package incentive

import (
	"time"

	"github.com/babylonchain/babylon/x/incentive/keeper"
	"github.com/babylonchain/babylon/x/incentive/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// handle coins in the fee collector account, including
	// - send a portion of coins in the fee collector account to the incentive module account
	// - accumulate BTC staking gauge at the current height
	// - accumulate BTC timestamping gauge at the current epoch
	k.HandleCoinsInFeeCollector(ctx)
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return []abci.ValidatorUpdate{}
}
