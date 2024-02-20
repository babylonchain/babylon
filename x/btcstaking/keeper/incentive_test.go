package keeper_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzRecordVotingPowerDistCache(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).AnyTimes()
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		keeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)

		// covenant and slashing addr
		covenantSKs, _, covenantQuorum := datagen.GenCovenantCommittee(r)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		slashingChangeLockTime := uint16(101)

		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// generate a random batch of finality providers
		numFpsWithVotingPower := datagen.RandomInt(r, 10) + 2
		numFps := numFpsWithVotingPower + datagen.RandomInt(r, 10)
		fpsWithVotingPowerMap := map[string]*types.FinalityProvider{}
		for i := uint64(0); i < numFps; i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			keeper.SetFinalityProvider(ctx, fp)
			if i < numFpsWithVotingPower {
				// these finality providers will receive BTC delegations and have voting power
				fpsWithVotingPowerMap[fp.BabylonPk.String()] = fp
			}
		}

		// for the first numFpsWithVotingPower finality providers, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		stakingValue := datagen.RandomInt(r, 100000) + 100000
		for _, fp := range fpsWithVotingPowerMap {
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(
					r,
					t,
					[]bbn.BIP340PubKey{*fp.BtcPk},
					delSK,
					covenantSKs,
					covenantQuorum,
					slashingAddress.EncodeAddress(),
					1, 1000, stakingValue,
					slashingRate,
					slashingChangeLockTime,
				)
				require.NoError(t, err)
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)
			}
		}

		// record voting power distribution cache
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		// assert voting power distribution cache is correct
		dc := keeper.GetVotingPowerDistCache(ctx, babylonHeight)
		require.NotNil(t, dc)
		require.Equal(t, dc.TotalVotingPower, numFpsWithVotingPower*numBTCDels*stakingValue)
		for _, fpDistInfo := range dc.TopFinalityProviders {
			require.Equal(t, fpDistInfo.TotalVotingPower, numBTCDels*stakingValue)
			fp, ok := fpsWithVotingPowerMap[fpDistInfo.BabylonPk.String()]
			require.True(t, ok)
			require.Equal(t, fpDistInfo.Commission, fp.Commission)
			require.Len(t, fpDistInfo.BtcDels, int(numBTCDels))
			for _, delDistInfo := range fpDistInfo.BtcDels {
				require.Equal(t, delDistInfo.VotingPower, stakingValue)
			}
		}
	})
}
