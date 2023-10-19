package btcstaking

import (
	"time"

	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// index BTC height at the current height
	k.IndexBTCHeight(ctx)
	// record voting power table at the current height
	k.RecordVotingPowerTable(ctx)
	// if BTC staking is activated, record reward distribution cache at the current height
	// TODO: consider merging RecordVotingPowerTable and RecordRewardDistCache so that we
	// only need to perform one full scan over BTC validators/delegations
	if k.IsBTCStakingActivated(ctx) {
		k.RecordRewardDistCache(ctx)
	}

}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return []abci.ValidatorUpdate{}
}
