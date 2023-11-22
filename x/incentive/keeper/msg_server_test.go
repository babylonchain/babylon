package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/incentive/keeper"
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := testkeeper.IncentiveKeeper(t, nil, nil, nil)
	return keeper.NewMsgServerImpl(*k), ctx
}

func TestMsgServer(t *testing.T) {
	ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
}

func FuzzWithdrawReward(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock bank keeper
		bk := types.NewMockBankKeeper(ctrl)

		ik, ctx := testkeeper.IncentiveKeeper(t, bk, nil, nil)
		ms := keeper.NewMsgServerImpl(*ik)

		// generate and set a random reward gauge with a random set of withdrawable coins
		rg := datagen.GenRandomRewardGauge(r)
		rg.WithdrawnCoins = datagen.GenRandomWithdrawnCoins(r, rg.Coins)
		sType := datagen.GenRandomStakeholderType(r)
		sAddr := datagen.GenRandomAccount().GetAddress()
		ik.SetRewardGauge(ctx, sType, sAddr, rg)

		// mock transfer of withdrawable coins
		withdrawableCoins := rg.GetWithdrawableCoins()
		bk.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Eq(types.ModuleName), gomock.Eq(sAddr), gomock.Eq(withdrawableCoins)).Times(1)

		// invoke withdraw and assert consistency
		resp, err := ms.WithdrawReward(ctx, &types.MsgWithdrawReward{
			Type:    sType.String(),
			Address: sAddr.String(),
		})
		require.NoError(t, err)
		require.Equal(t, withdrawableCoins, resp.Coins)

		// ensure reward gauge is now empty
		newRg := ik.GetRewardGauge(ctx, sType, sAddr)
		require.NotNil(t, newRg)
		require.True(t, newRg.IsFullyWithdrawn())
	})
}
