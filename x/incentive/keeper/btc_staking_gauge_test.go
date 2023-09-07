package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzRewardBTCStaking(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock bank keeper
		bankKeeper := types.NewMockBankKeeper(ctrl)

		// create incentive keeper
		keeper, ctx := testkeeper.IncentiveKeeper(t, bankKeeper, nil, nil)
		height := datagen.RandomInt(r, 1000)
		ctx = ctx.WithBlockHeight(int64(height))

		// set a random gauge
		gauge := datagen.GenRandomGauge(r)
		keeper.SetBTCStakingGauge(ctx, height, gauge)

		// generate a random reward distribution cache
		rdc, err := datagen.GenRandomRewardDistCache(r)
		require.NoError(t, err)

		// mock transfer to ensure reward is distributed correctly
		distributedCoins := sdk.NewCoins()
		for _, btcVal := range rdc.BtcVals {
			btcValPortion := rdc.GetBTCValPortion(btcVal)
			coinsForBTCValAndDels := gauge.GetCoinsPortion(btcValPortion)
			coinsForCommission := types.GetCoinsPortion(coinsForBTCValAndDels, *btcVal.Commission)
			if coinsForCommission.IsAllPositive() {
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(btcVal.GetAddress()), gomock.Eq(coinsForCommission)).Return(nil).Times(1)
				distributedCoins = distributedCoins.Add(coinsForCommission...)
			}
			coinsForBTCDels := coinsForBTCValAndDels.Sub(coinsForCommission...)
			for _, btcDel := range btcVal.BtcDels {
				btcDelPortion := btcVal.GetBTCDelPortion(btcDel)
				coinsForDel := types.GetCoinsPortion(coinsForBTCDels, btcDelPortion)
				if coinsForDel.IsAllPositive() {
					bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(btcDel.GetAddress()), gomock.Eq(coinsForDel)).Return(nil).Times(1)
					distributedCoins = distributedCoins.Add(coinsForDel...)
				}
			}
		}

		// distribute rewards in the gauge to BTC validators/delegations
		keeper.RewardBTCStaking(ctx, height, rdc)

		// assert distributedCoins is a subset of coins in gauge
		require.True(t, gauge.Coins.IsAllGTE(distributedCoins))
	})
}
