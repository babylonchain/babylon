package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
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
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		ckptKeeper := types.NewMockCheckpointingKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper, ckptKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		h.NoError(err)

		// generate a random batch of finality providers
		numFpsWithVotingPower := datagen.RandomInt(r, 10) + 2
		numFps := numFpsWithVotingPower + datagen.RandomInt(r, 10)
		fpsWithVotingPowerMap := map[string]*types.FinalityProvider{}
		for i := uint64(0); i < numFps; i++ {
			_, _, fp := h.CreateFinalityProvider(r)
			if i < numFpsWithVotingPower {
				// these finality providers will receive BTC delegations and have voting power
				fpsWithVotingPowerMap[fp.Addr] = fp
			}
		}

		// for the first numFpsWithVotingPower finality providers, generate a random number of BTC delegations and add covenant signatures to activate them
		numBTCDels := datagen.RandomInt(r, 10) + 1
		stakingValue := datagen.RandomInt(r, 100000) + 100000
		for _, fp := range fpsWithVotingPowerMap {
			for j := uint64(0); j < numBTCDels; j++ {
				_, _, _, delMsg, del := h.CreateDelegation(
					r,
					fp.BtcPk.MustToBTCPK(),
					changeAddress.EncodeAddress(),
					int64(stakingValue),
					1000,
				)
				h.CreateCovenantSigs(r, covenantSKs, delMsg, del)
			}
		}

		// record voting power distribution cache
		babylonHeight := datagen.RandomInt(r, 10) + 1
		h.Ctx = datagen.WithCtxHeight(h.Ctx, babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		require.NoError(t, err)

		// assert voting power distribution cache is correct
		dc, err := h.BTCStakingKeeper.GetVotingPowerDistCache(h.Ctx, babylonHeight)
		require.NoError(t, err)
		require.NotNil(t, dc)
		require.Equal(t, dc.TotalVotingPower, numFpsWithVotingPower*numBTCDels*stakingValue)
		maxNumFps := h.BTCStakingKeeper.GetParams(h.Ctx).MaxActiveFinalityProviders
		activeFPs := dc.GetActiveFinalityProviders(maxNumFps)
		for _, fpDistInfo := range activeFPs {
			require.Equal(t, fpDistInfo.TotalVotingPower, numBTCDels*stakingValue)
			fp, ok := fpsWithVotingPowerMap[fpDistInfo.Addr]
			require.True(t, ok)
			require.Equal(t, fpDistInfo.Commission, fp.Commission)
			require.Len(t, fpDistInfo.BtcDels, int(numBTCDels))
			for _, delDistInfo := range fpDistInfo.BtcDels {
				require.Equal(t, delDistInfo.VotingPower, stakingValue)
			}
		}
	})
}
