package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzVotingPowerTable(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		keeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)

		// generate a random batch of validators
		btcVals := []*types.BTCValidator{}
		numBTCValsWithVotingPower := datagen.RandomInt(r, 10) + 1
		numBTCVals := numBTCValsWithVotingPower + datagen.RandomInt(r, 10)
		for i := uint64(0); i < numBTCVals; i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)
			btcVals = append(btcVals, btcVal)
		}

		// for the first numBTCValsWithVotingPower validators, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			valBTCPK := btcVals[i].BtcPk
			for j := uint64(0); j < numBTCDels; j++ {
				btcDel, err := datagen.GenRandomBTCDelegation(r, valBTCPK, 1, 1000, 1) // timelock period: 1-1000
				require.NoError(t, err)
				keeper.SetBTCDelegation(ctx, btcDel)
			}
		}

		// Case 1: assert none of validators has voting power (since BTC height is 0)
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 0}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)
		for _, btcVal := range btcVals {
			power := keeper.GetVotingPower(ctx, *btcVal.BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// since there is no BTC validator with BTC delegation, the BTC staking protocol is not activated yet
		_, err := keeper.GetBTCStakingActivatedHeight(ctx)
		require.Error(t, err)

		// Case 2: move to 1st BTC block, then assert the first numBTCValsWithVotingPower validators have voting power
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			require.Equal(t, uint64(numBTCDels), power)
		}
		for i := numBTCValsWithVotingPower; i < numBTCVals; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// also, get voting power table and assert consistency
		powerTable := keeper.GetVotingPowerTable(ctx, babylonHeight)
		require.NotNil(t, powerTable)
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			require.Equal(t, powerTable[btcVals[i].BtcPk.ToHexStr()], power)
		}
		// the activation height should be the current Babylon height as well
		activatedHeight, err := keeper.GetBTCStakingActivatedHeight(ctx)
		require.NoError(t, err)
		require.Equal(t, babylonHeight, activatedHeight)

		// Case 3: move to 999th BTC block, then assert none of validators has voting power (since end height - w < BTC height)
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 999}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)
		for _, btcVal := range btcVals {
			power := keeper.GetVotingPower(ctx, *btcVal.BtcPk, babylonHeight)
			require.Zero(t, power)
		}

	})
}
