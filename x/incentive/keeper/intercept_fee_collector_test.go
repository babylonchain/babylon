package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	feeCollectorAcc = authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	fees            = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
)

func FuzzInterceptFeeCollector(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock bank keeper
		bankKeeper := types.NewMockBankKeeper(ctrl)
		bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(fees).Times(1)

		// mock account keeper
		accountKeeper := types.NewMockAccountKeeper(ctrl)
		accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), authtypes.FeeCollectorName).Return(feeCollectorAcc).Times(1)

		// mock epoching keeper
		epochNum := datagen.RandomInt(r, 100) + 1
		epochingKeeper := types.NewMockEpochingKeeper(ctrl)
		epochingKeeper.EXPECT().GetEpoch(gomock.Any()).Return(&epochingtypes.Epoch{EpochNumber: epochNum}).Times(1)

		keeper, ctx := testkeeper.IncentiveKeeper(t, bankKeeper, accountKeeper, epochingKeeper)
		height := datagen.RandomInt(r, 1000)
		ctx = ctx.WithBlockHeight(int64(height))

		// mock (thus ensure) that fees with the exact portion is intercepted
		// NOTE: if the actual fees are different from feesForIncentive the test will fail
		params := keeper.GetParams(ctx)
		feesForBTCStaking := types.GetCoinsPortion(fees, params.BTCStakingPortion())
		feesForBTCTimestamping := types.GetCoinsPortion(fees, params.BTCTimestampingPortion())
		bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Eq(authtypes.FeeCollectorName), gomock.Eq(types.ModuleName), gomock.Eq(feesForBTCStaking)).Times(1)
		bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Eq(authtypes.FeeCollectorName), gomock.Eq(types.ModuleName), gomock.Eq(feesForBTCTimestamping)).Times(1)

		// handle coins in fee collector
		keeper.HandleCoinsInFeeCollector(ctx)

		// assert correctness of BTC staking gauge at height
		btcStakingFee := types.GetCoinsPortion(fees, params.BTCStakingPortion())
		btcStakingGauge := keeper.GetBTCStakingGauge(ctx, height)
		require.NotNil(t, btcStakingGauge)
		require.Equal(t, btcStakingFee, btcStakingGauge.Coins)

		// assert correctness of BTC timestamping gauge at epoch
		btcTimestampingFee := types.GetCoinsPortion(fees, params.BTCTimestampingPortion())
		btcTimestampingGauge := keeper.GetBTCTimestampingGauge(ctx, epochNum)
		require.NotNil(t, btcTimestampingGauge)
		require.Equal(t, btcTimestampingFee, btcTimestampingGauge.Coins)

		// accumulate for this epoch again and see if the epoch's BTC timestamping gauge has accumulated or not
		height += 1
		ctx = ctx.WithBlockHeight(int64(height))
		bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(fees).Times(1)
		accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), authtypes.FeeCollectorName).Return(feeCollectorAcc).Times(1)
		epochingKeeper.EXPECT().GetEpoch(gomock.Any()).Return(&epochingtypes.Epoch{EpochNumber: epochNum}).Times(1)
		bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Eq(authtypes.FeeCollectorName), gomock.Eq(types.ModuleName), gomock.Eq(feesForBTCStaking)).Times(1)
		bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Eq(authtypes.FeeCollectorName), gomock.Eq(types.ModuleName), gomock.Eq(feesForBTCTimestamping)).Times(1)
		// handle coins in fee collector
		keeper.HandleCoinsInFeeCollector(ctx)
		// assert BTC timestamping gauge has doubled
		btcTimestampingGauge2 := keeper.GetBTCTimestampingGauge(ctx, epochNum)
		require.NotNil(t, btcTimestampingGauge2)
		for i := range btcTimestampingGauge.Coins {
			amount := btcTimestampingGauge.Coins[i].Amount.Uint64()
			amount2 := btcTimestampingGauge2.Coins[i].Amount.Uint64()
			require.Equal(t, amount*2, amount2)
		}
	})
}
