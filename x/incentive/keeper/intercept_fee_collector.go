package keeper

import (
	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HandleCoinsInFeeCollector intercepts a portion of coins in fee collector, and distributes
// them to BTC staking gauge and BTC timestamping gauge of the current height and epoch, respectively.
// It is invoked upon every `BeginBlock`.
// adapted from https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/x/distribution/keeper/allocation.go#L15-L26
func (k Keeper) HandleCoinsInFeeCollector(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// find the fee collector account
	feeCollector := k.accountKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	// get all balances in the fee collector account,
	// where the balance includes minted tokens in the previous block
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())

	// don't intercept if there is no fee in fee collector account
	if !feesCollectedInt.IsAllPositive() {
		return
	}

	// record BTC staking gauge for the current height, and transfer corresponding amount
	// from fee collector account to incentive module account
	// TODO: maybe we should not transfer reward to BTC staking gauge before BTC staking is activated
	// this is tricky to implement since finality module will depend on incentive and incentive cannot
	// depend on finality module due to cyclic dependency
	btcStakingPortion := params.BTCStakingPortion()
	btcStakingReward := types.GetCoinsPortion(feesCollectedInt, btcStakingPortion)
	k.accumulateBTCStakingReward(ctx, btcStakingReward)

	// record BTC timestamping gauge for the current epoch, and transfer corresponding amount
	// from fee collector account to incentive module account
	btcTimestampingPortion := params.BTCTimestampingPortion()
	btcTimestampingReward := types.GetCoinsPortion(feesCollectedInt, btcTimestampingPortion)
	k.accumulateBTCTimestampingReward(ctx, btcTimestampingReward)
}
