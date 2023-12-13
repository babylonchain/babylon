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

	// index BTC height at the current height
	k.IndexBTCHeight(ctx)
	// record voting power table at the current height
	k.RecordVotingPowerTable(ctx)
	// if BTC staking is activated, record reward distribution cache at the current height
	// TODO: consider merging RecordVotingPowerTable and RecordRewardDistCache so that we
	// only need to perform one full scan over finality providers/delegations
	if k.IsBTCStakingActivated(ctx) {
		k.RecordRewardDistCache(ctx)
	}
	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return []abci.ValidatorUpdate{}, nil
}
