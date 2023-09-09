package keeper_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzRewardBTCTimestamping(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock bank keeper
		bankKeeper := types.NewMockBankKeeper(ctrl)

		// create incentive keeper
		keeper, ctx := testkeeper.IncentiveKeeper(t, bankKeeper, nil, nil)
		epoch := datagen.RandomInt(r, 1000) + 1

		// set a random gauge
		gauge := datagen.GenRandomGauge(r)
		keeper.SetBTCTimestampingGauge(ctx, epoch, gauge)

		// generate a random BTC timestamping reward distribution info
		rdi := datagen.GenRandomBTCTimestampingRewardDistInfo(r)

		// parameters
		params := types.DefaultParams()
		btcTimestampingPortion := params.BTCTimestampingPortion()
		bestPortion := math.LegacyNewDecWithPrec(80, 2) // 80 * 10^{-2} = 0.8

		// expected values
		distributedCoins := sdk.NewCoins()
		submitterRewardMap := map[string]sdk.Coins{} // key: address, value: reward
		reporterRewardMap := map[string]sdk.Coins{}  // key: address, value: reward

		submitterPortion := params.SubmitterPortion.QuoTruncate(btcTimestampingPortion)
		coinsToSubmitters := gauge.GetCoinsPortion(submitterPortion)
		coinsToBestSubmitter := types.GetCoinsPortion(coinsToSubmitters, bestPortion)
		if coinsToBestSubmitter.IsAllPositive() {
			submitterRewardMap[rdi.Best.Submitter.String()] = coinsToBestSubmitter
			distributedCoins.Add(coinsToBestSubmitter...)
		}
		// best reporter
		reporterPortion := params.ReporterPortion.QuoTruncate(btcTimestampingPortion)
		coinsToReporters := gauge.GetCoinsPortion(reporterPortion)
		coinsToBestReporter := types.GetCoinsPortion(coinsToReporters, bestPortion)
		if coinsToBestReporter.IsAllPositive() {
			reporterRewardMap[rdi.Best.Reporter.String()] = coinsToBestReporter
			distributedCoins.Add(coinsToBestReporter...)
		}
		// other submitters and reporters
		if len(rdi.Others) > 0 {
			// other submitters
			coinsToOtherSubmitters := coinsToSubmitters.Sub(coinsToBestSubmitter...)
			eachOtherSubmitterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
			coinsToEachOtherSubmitter := types.GetCoinsPortion(coinsToOtherSubmitters, eachOtherSubmitterPortion)
			if coinsToEachOtherSubmitter.IsAllPositive() {
				for _, submission := range rdi.Others {
					submitterRewardMap[submission.Submitter.String()] = coinsToEachOtherSubmitter
					distributedCoins.Add(coinsToEachOtherSubmitter...)
				}
			}
			// other reporters
			coinsToOtherReporters := coinsToReporters.Sub(coinsToBestReporter...)
			eachOtherReporterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
			coinsToEachOtherReporter := types.GetCoinsPortion(coinsToOtherReporters, eachOtherReporterPortion)
			if coinsToEachOtherReporter.IsAllPositive() {
				for _, submission := range rdi.Others {
					reporterRewardMap[submission.Reporter.String()] = coinsToEachOtherReporter
					distributedCoins.Add(coinsToEachOtherReporter...)
				}
			}
		} else {
			// no other submission. give rest coins to best submitter/reporter
			// give rest coins to the best submitter
			restCoinsToSubmitter := coinsToSubmitters.Sub(coinsToBestSubmitter...)
			if restCoinsToSubmitter.IsAllPositive() {
				submitterRewardMap[rdi.Best.Submitter.String()] = submitterRewardMap[rdi.Best.Submitter.String()].Add(restCoinsToSubmitter...)
				distributedCoins.Add(restCoinsToSubmitter...)
			}
			// give rest coins to the best reporter
			restCoinsToReporter := coinsToReporters.Sub(coinsToBestReporter...)
			if restCoinsToReporter.IsAllPositive() {
				reporterRewardMap[rdi.Best.Reporter.String()] = reporterRewardMap[rdi.Best.Reporter.String()].Add(restCoinsToSubmitter...)
				distributedCoins.Add(restCoinsToReporter...)
			}
		}

		// distribute rewards in the gauge to BTC validators/delegations
		keeper.RewardBTCTimestamping(ctx, epoch, rdi)

		// assert consistency between reward map and reward gauge
		for addrStr, reward := range submitterRewardMap {
			addr, err := sdk.AccAddressFromBech32(addrStr)
			require.NoError(t, err)
			rg := keeper.GetRewardGauge(ctx, types.SubmitterType, addr)
			require.NotNil(t, rg)
			require.Equal(t, reward, rg.Coins)
		}
		for addrStr, reward := range reporterRewardMap {
			addr, err := sdk.AccAddressFromBech32(addrStr)
			require.NoError(t, err)
			rg := keeper.GetRewardGauge(ctx, types.ReporterType, addr)
			require.NotNil(t, rg)
			require.Equal(t, reward, rg.Coins)
		}

		// assert distributedCoins is a subset of coins in gauge
		require.True(t, gauge.Coins.IsAllGTE(distributedCoins))
	})
}
