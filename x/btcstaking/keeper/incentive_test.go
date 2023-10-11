package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzRecordRewardDistCache(f *testing.F) {
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

		// jury and slashing addr
		jurySK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		// generate a random batch of validators
		numBTCValsWithVotingPower := datagen.RandomInt(r, 10) + 2
		numBTCVals := numBTCValsWithVotingPower + datagen.RandomInt(r, 10)
		btcValsWithVotingPowerMap := map[string]*types.BTCValidator{}
		for i := uint64(0); i < numBTCVals; i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)
			if i < numBTCValsWithVotingPower {
				// these BTC validators will receive BTC delegations and have voting power
				btcValsWithVotingPowerMap[btcVal.BabylonPk.String()] = btcVal
			}
		}

		// for the first numBTCValsWithVotingPower validators, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		stakingValue := datagen.RandomInt(r, 100000) + 100000
		for _, btcVal := range btcValsWithVotingPowerMap {
			valBTCPK := btcVal.BtcPk
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(r, valBTCPK, delSK, jurySK, slashingAddr.String(), 1, 1000, stakingValue) // timelock period: 1-1000
				require.NoError(t, err)
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)
			}
		}

		// record reward distribution cache
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordRewardDistCache(ctx)

		// assert reward distribution cache is correct
		rdc, err := keeper.GetRewardDistCache(ctx, babylonHeight)
		require.NoError(t, err)
		require.Equal(t, rdc.TotalVotingPower, numBTCValsWithVotingPower*numBTCDels*stakingValue)
		for _, valDistInfo := range rdc.BtcVals {
			require.Equal(t, valDistInfo.TotalVotingPower, numBTCDels*stakingValue)
			btcVal, ok := btcValsWithVotingPowerMap[valDistInfo.BabylonPk.String()]
			require.True(t, ok)
			require.Equal(t, valDistInfo.Commission, btcVal.Commission)
			require.Len(t, valDistInfo.BtcDels, int(numBTCDels))
			for _, delDistInfo := range valDistInfo.BtcDels {
				require.Equal(t, delDistInfo.VotingPower, stakingValue)
			}
		}
	})
}
