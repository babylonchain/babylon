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
		ctx = datagen.WithCtxHeight(ctx, height)

		// set a random gauge
		gauge := datagen.GenRandomGauge(r)
		keeper.SetBTCStakingGauge(ctx, height, gauge)

		// generate a random voting power distribution cache
		dc, err := datagen.GenRandomVotingPowerDistCache(r, 100)
		require.NoError(t, err)

		// expected values
		distributedCoins := sdk.NewCoins()
		fpRewardMap := map[string]sdk.Coins{}     // key: address, value: reward
		btcDelRewardMap := map[string]sdk.Coins{} // key: address, value: reward

		for _, fp := range dc.ActiveFinalityProviders {
			fpPortion := dc.GetFinalityProviderPortion(fp)
			coinsForFpsAndDels := gauge.GetCoinsPortion(fpPortion)
			coinsForCommission := types.GetCoinsPortion(coinsForFpsAndDels, *fp.Commission)
			if coinsForCommission.IsAllPositive() {
				fpRewardMap[fp.GetAddress().String()] = coinsForCommission
				distributedCoins.Add(coinsForCommission...)
			}
			coinsForBTCDels := coinsForFpsAndDels.Sub(coinsForCommission...)
			for _, btcDel := range fp.BtcDels {
				btcDelPortion := fp.GetBTCDelPortion(btcDel)
				coinsForDel := types.GetCoinsPortion(coinsForBTCDels, btcDelPortion)
				if coinsForDel.IsAllPositive() {
					btcDelRewardMap[btcDel.GetAddress().String()] = coinsForDel
					distributedCoins.Add(coinsForDel...)
				}
			}
		}

		// distribute rewards in the gauge to finality providers/delegations
		keeper.RewardBTCStaking(ctx, height, dc)

		// assert consistency between reward map and reward gauge
		for addrStr, reward := range fpRewardMap {
			addr, err := sdk.AccAddressFromBech32(addrStr)
			require.NoError(t, err)
			rg := keeper.GetRewardGauge(ctx, types.FinalityProviderType, addr)
			require.NotNil(t, rg)
			require.Equal(t, reward, rg.Coins)
		}
		for addrStr, reward := range btcDelRewardMap {
			addr, err := sdk.AccAddressFromBech32(addrStr)
			require.NoError(t, err)
			rg := keeper.GetRewardGauge(ctx, types.BTCDelegationType, addr)
			require.NotNil(t, rg)
			require.Equal(t, reward, rg.Coins)
		}

		// assert distributedCoins is a subset of coins in gauge
		require.True(t, gauge.Coins.IsAllGTE(distributedCoins))
	})
}
