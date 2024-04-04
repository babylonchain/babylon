package keeper_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
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
		ckptKeeper := types.NewMockCheckpointingKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper, ckptKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		h.NoError(err)

		// generate a random batch of finality providers
		fps := []*types.FinalityProvider{}
		numFpsWithVotingPower := datagen.RandomInt(r, 10) + 2
		numFps := numFpsWithVotingPower + datagen.RandomInt(r, 10)
		for i := uint64(0); i < numFps; i++ {
			_, _, fp := h.CreateFinalityProvider(r)
			fps = append(fps, fp)
		}

		// mock that the registered epoch is finalised
		h.CheckpointingKeeper.EXPECT().GetLastFinalizedEpoch(gomock.Any()).Return(uint64(10)).AnyTimes()

		// for the first numFpsWithVotingPower finality providers, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		stakingValue := datagen.RandomInt(r, 100000) + 100000
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			for j := uint64(0); j < numBTCDels; j++ {
				_, _, _, delMsg, del := h.CreateDelegation(
					r,
					fps[i].BtcPk.MustToBTCPK(),
					changeAddress.EncodeAddress(),
					int64(stakingValue),
					1000,
				)
				h.CreateCovenantSigs(r, covenantSKs, delMsg, del)
			}
		}

		/*
			assert the first numFpsWithVotingPower finality providers have voting power
		*/
		babylonHeight := datagen.RandomInt(r, 10) + 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		require.NoError(t, err)

		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fps[i].BtcPk, babylonHeight)
			require.Equal(t, numBTCDels*stakingValue, power)
		}
		for i := numFpsWithVotingPower; i < numFps; i++ {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fps[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// also, get voting power table and assert consistency
		powerTable := h.BTCStakingKeeper.GetVotingPowerTable(h.Ctx, babylonHeight)
		require.NotNil(t, powerTable)
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fps[i].BtcPk, babylonHeight)
			require.Equal(t, powerTable[fps[i].BtcPk.MarshalHex()], power)
		}
		// the activation height should be the current Babylon height as well
		activatedHeight, err := h.BTCStakingKeeper.GetBTCStakingActivatedHeight(h.Ctx)
		require.NoError(t, err)
		require.Equal(t, babylonHeight, activatedHeight)

		/*
			slash a random finality provider and move on
			then assert the slashed finality provider does not have voting power
		*/
		// move to next Babylon height
		h.BTCLightClientKeeper = btclcKeeper
		babylonHeight += 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		// slash a random finality provider
		slashedIdx := datagen.RandomInt(r, int(numFpsWithVotingPower))
		slashedFp := fps[slashedIdx]
		err = h.BTCStakingKeeper.SlashFinalityProvider(h.Ctx, slashedFp.BtcPk.MustMarshal())
		require.NoError(t, err)
		// index height and record power table
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		require.NoError(t, err)

		// check if the slashed finality provider's voting power becomes zero
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fps[i].BtcPk, babylonHeight)
			if i == slashedIdx {
				require.Zero(t, power)
			} else {
				require.Equal(t, numBTCDels*stakingValue, power)
			}
		}
		for i := numFpsWithVotingPower; i < numFps; i++ {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fps[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// also, get voting power table and assert consistency
		powerTable = h.BTCStakingKeeper.GetVotingPowerTable(h.Ctx, babylonHeight)
		require.NotNil(t, powerTable)
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fps[i].BtcPk, babylonHeight)
			if i == slashedIdx {
				require.Zero(t, power)
			}
			require.Equal(t, powerTable[fps[i].BtcPk.MarshalHex()], power)
		}

		/*
			move to 999th BTC block, then assert none of finality providers has voting power (since end height - w < BTC height)
		*/
		// replace the old mocked keeper
		babylonHeight += 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 999}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		require.NoError(t, err)

		for _, fp := range fps {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fp.BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// the activation height should be same as before
		activatedHeight2, err := h.BTCStakingKeeper.GetBTCStakingActivatedHeight(h.Ctx)
		require.NoError(t, err)
		require.Equal(t, activatedHeight, activatedHeight2)
	})
}

func FuzzVotingPowerTable_ActiveFinalityProviders(f *testing.F) {
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

		// mock that the registered epoch is finalised
		h.CheckpointingKeeper.EXPECT().GetLastFinalizedEpoch(gomock.Any()).Return(uint64(10)).AnyTimes()

		// generate a random batch of finality providers, each with a BTC delegation with random power
		fpsWithMeta := []*types.FinalityProviderDistInfo{}
		numFps := datagen.RandomInt(r, 300) + 1
		for i := uint64(0); i < numFps; i++ {
			// generate finality provider
			_, _, fp := h.CreateFinalityProvider(r)

			// delegate to this finality provider
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			_, _, _, delMsg, del := h.CreateDelegation(
				r,
				fp.BtcPk.MustToBTCPK(),
				changeAddress.EncodeAddress(),
				int64(stakingValue),
				1000,
			)
			h.CreateCovenantSigs(r, covenantSKs, delMsg, del)

			// record voting power
			fpsWithMeta = append(fpsWithMeta, &types.FinalityProviderDistInfo{
				BtcPk:            fp.BtcPk,
				TotalVotingPower: stakingValue,
			})
		}

		maxActiveFpsParam := h.BTCStakingKeeper.GetParams(h.Ctx).MaxActiveFinalityProviders
		// get a map of expected active finality providers
		types.SortFinalityProviders(fpsWithMeta)
		expectedActiveFps := fpsWithMeta[:min(uint32(len(fpsWithMeta)), maxActiveFpsParam)]
		expectedActiveFpsMap := map[string]uint64{}
		for _, fp := range expectedActiveFps {
			expectedActiveFpsMap[fp.BtcPk.MarshalHex()] = fp.TotalVotingPower
		}

		// record voting power table
		babylonHeight := datagen.RandomInt(r, 10) + 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		require.NoError(t, err)

		//  only finality providers in expectedActiveFpsMap have voting power
		for _, fp := range fpsWithMeta {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, fp.BtcPk.MustMarshal(), babylonHeight)
			if expectedPower, ok := expectedActiveFpsMap[fp.BtcPk.MarshalHex()]; ok {
				require.Equal(t, expectedPower, power)
			} else {
				require.Zero(t, power)
			}
		}

		// also, get voting power table and assert there is
		// min(len(expectedActiveFps), MaxActiveFinalityProviders) active finality providers
		powerTable := h.BTCStakingKeeper.GetVotingPowerTable(h.Ctx, babylonHeight)
		expectedNumActiveFps := len(expectedActiveFpsMap)
		if expectedNumActiveFps > int(maxActiveFpsParam) {
			expectedNumActiveFps = int(maxActiveFpsParam)
		}
		require.Len(t, powerTable, expectedNumActiveFps)
		// assert consistency of voting power
		for pkHex, expectedPower := range expectedActiveFpsMap {
			require.Equal(t, powerTable[pkHex], expectedPower)
		}
	})
}

func FuzzVotingPowerTable_ActiveFinalityProviderRotation(f *testing.F) {
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
		// set random number of max number of finality providers
		// in order to cover cases that number of finality providers is more or
		// less than `MaxActiveFinalityProviders`
		bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
		bsParams.MaxActiveFinalityProviders = uint32(datagen.RandomInt(r, 20) + 10)
		err := h.BTCStakingKeeper.SetParams(h.Ctx, bsParams)
		h.NoError(err)
		// change address
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		h.NoError(err)

		// mock that the registered epoch is finalised
		h.CheckpointingKeeper.EXPECT().GetLastFinalizedEpoch(gomock.Any()).Return(uint64(10)).AnyTimes()

		numFps := datagen.RandomInt(r, 20) + 10
		numActiveFPs := int(min(numFps, uint64(bsParams.MaxActiveFinalityProviders)))

		/*
			Generate a random batch of finality providers, each with a BTC delegation
			with random voting power.
			Then, assert voting power table
		*/
		fpsWithMeta := []*types.FinalityProviderWithMeta{}
		for i := uint64(0); i < numFps; i++ {
			// generate finality provider
			// generate and insert new finality provider
			_, fpPK, fp := h.CreateFinalityProvider(r)

			// create BTC delegation and add covenant signatures to activate it
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			_, _, _, delMsg, del := h.CreateDelegation(
				r,
				fpPK,
				changeAddress.EncodeAddress(),
				int64(stakingValue),
				1000,
			)
			h.CreateCovenantSigs(r, covenantSKs, delMsg, del)

			// record voting power
			fpsWithMeta = append(fpsWithMeta, &types.FinalityProviderWithMeta{
				BtcPk:       fp.BtcPk,
				VotingPower: stakingValue,
			})
		}

		// record voting power table
		babylonHeight := datagen.RandomInt(r, 10) + 1
		h.Ctx = datagen.WithCtxHeight(h.Ctx, babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)

		// assert that only top `min(MaxActiveFinalityProviders, numFPs)` finality providers have voting power
		sort.SliceStable(fpsWithMeta, func(i, j int) bool {
			return fpsWithMeta[i].VotingPower > fpsWithMeta[j].VotingPower
		})
		for i := 0; i < numActiveFPs; i++ {
			votingPower := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fpsWithMeta[i].BtcPk, babylonHeight)
			require.Equal(t, fpsWithMeta[i].VotingPower, votingPower)
		}
		for i := numActiveFPs; i < int(numFps); i++ {
			votingPower := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fpsWithMeta[i].BtcPk, babylonHeight)
			require.Zero(t, votingPower)
		}

		/*
			Delegate more tokens to some existing finality providers
			, and create some new finality providers
			Then assert voting power table again
		*/
		// delegate more tokens to some existing finality providers
		for i := uint64(0); i < numFps; i++ {
			if !datagen.OneInN(r, 2) {
				continue
			}

			stakingValue := datagen.RandomInt(r, 100000) + 100000
			fpBTCPK := fpsWithMeta[i].BtcPk
			_, _, _, delMsg, del := h.CreateDelegation(
				r,
				fpBTCPK.MustToBTCPK(),
				changeAddress.EncodeAddress(),
				int64(stakingValue),
				1000,
			)
			h.CreateCovenantSigs(r, covenantSKs, delMsg, del)

			// accumulate voting power for this finality provider
			fpsWithMeta[i].VotingPower += stakingValue

			break
		}
		// create more finality providers
		numNewFps := datagen.RandomInt(r, 20) + 10
		numFps += numNewFps
		numActiveFPs = int(min(numFps, uint64(bsParams.MaxActiveFinalityProviders)))
		for i := uint64(0); i < numNewFps; i++ {
			// generate finality provider
			// generate and insert new finality provider
			_, fpPK, fp := h.CreateFinalityProvider(r)

			// create BTC delegation and add covenant signatures to activate it
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			_, _, _, delMsg, del := h.CreateDelegation(
				r,
				fpPK,
				changeAddress.EncodeAddress(),
				int64(stakingValue),
				1000,
			)
			h.CreateCovenantSigs(r, covenantSKs, delMsg, del)

			// record voting power
			fpsWithMeta = append(fpsWithMeta, &types.FinalityProviderWithMeta{
				BtcPk:       fp.BtcPk,
				VotingPower: stakingValue,
			})
		}

		// record voting power table
		babylonHeight += 1
		h.Ctx = datagen.WithCtxHeight(h.Ctx, babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)

		// again, assert that only top `min(MaxActiveFinalityProviders, numFPs)` finality providers have voting power
		sort.SliceStable(fpsWithMeta, func(i, j int) bool {
			return fpsWithMeta[i].VotingPower > fpsWithMeta[j].VotingPower
		})
		for i := 0; i < numActiveFPs; i++ {
			votingPower := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fpsWithMeta[i].BtcPk, babylonHeight)
			require.Equal(t, fpsWithMeta[i].VotingPower, votingPower)
		}
		for i := numActiveFPs; i < int(numFps); i++ {
			votingPower := h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fpsWithMeta[i].BtcPk, babylonHeight)
			require.Zero(t, votingPower)
		}
	})
}
