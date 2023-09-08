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

		// mock bank transfer here
		// best submitter
		distributedCoins := sdk.NewCoins()
		submitterPortion := params.SubmitterPortion.QuoTruncate(btcTimestampingPortion)
		coinsToSubmitters := gauge.GetCoinsPortion(submitterPortion)
		coinsToBestSubmitter := types.GetCoinsPortion(coinsToSubmitters, bestPortion)
		if coinsToBestSubmitter.IsAllPositive() {
			bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(rdi.Best.Submitter), gomock.Eq(coinsToBestSubmitter)).Times(1)
			distributedCoins.Add(coinsToBestSubmitter...)
		}
		// best reporter
		reporterPortion := params.ReporterPortion.QuoTruncate(btcTimestampingPortion)
		coinsToReporters := gauge.GetCoinsPortion(reporterPortion)
		coinsToBestReporter := types.GetCoinsPortion(coinsToReporters, bestPortion)
		if coinsToBestReporter.IsAllPositive() {
			bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(rdi.Best.Reporter), gomock.Eq(coinsToBestReporter)).Times(1)
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
					bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(submission.Submitter), gomock.Eq(coinsToEachOtherSubmitter)).Times(1)
				}
				distributedCoins.Add(coinsToEachOtherSubmitter...)
			}
			// other reporters
			coinsToOtherReporters := coinsToReporters.Sub(coinsToBestReporter...)
			eachOtherReporterPortion := math.LegacyOneDec().QuoTruncate(math.LegacyOneDec().MulInt64(int64(len(rdi.Others))))
			coinsToEachOtherReporter := types.GetCoinsPortion(coinsToOtherReporters, eachOtherReporterPortion)
			if coinsToEachOtherReporter.IsAllPositive() {
				for _, submission := range rdi.Others {
					bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(submission.Reporter), gomock.Eq(coinsToEachOtherReporter)).Times(1)
				}
				distributedCoins.Add(coinsToEachOtherReporter...)
			}
		} else {
			// no other submission. give rest coins to best submitter/reporter
			// give rest coins to the best submitter
			restCoinsToSubmitter := coinsToSubmitters.Sub(coinsToBestSubmitter...)
			if restCoinsToSubmitter.IsAllPositive() {
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(rdi.Best.Submitter), gomock.Eq(restCoinsToSubmitter)).Times(1)
			}
			// give rest coins to the best reporter
			restCoinsToReporter := coinsToReporters.Sub(coinsToBestReporter...)
			if restCoinsToReporter.IsAllPositive() {
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(rdi.Best.Reporter), gomock.Eq(restCoinsToReporter)).Times(1)
			}
		}

		// distribute rewards in the gauge to BTC validators/delegations
		keeper.RewardBTCTimestamping(ctx, epoch, rdi)

		// assert distributedCoins is a subset of coins in gauge
		require.True(t, gauge.Coins.IsAllGTE(distributedCoins))
	})
}
