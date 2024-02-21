package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
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
		h := NewHelper(t, btclcKeeper, btccKeeper)

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
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		h.NoError(err)

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
		expectedActiveFps := types.FilterTopNFinalityProviders(fpsWithMeta, maxActiveFpsParam)
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
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		h.NoError(err)

		// generate a random batch of finality providers, each with a BTC delegation with random power
		fpsWithMeta := []*types.FinalityProviderWithMeta{}
		numFps := uint64(200) // there has to be more than `maxActiveFinalityProviders` finality providers
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

		// get maps of active/inactive finality providers
		activeFpsMap := map[string]uint64{}
		inactiveFpsMap := map[string]uint64{}
		for _, fp := range fpsWithMeta {
			power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, fp.BtcPk.MustMarshal(), babylonHeight)
			if power > 0 {
				activeFpsMap[fp.BtcPk.MarshalHex()] = power
			} else {
				inactiveFpsMap[fp.BtcPk.MarshalHex()] = power
			}
		}

		// delegate a huge amount of tokens to one of the inactive finality provider
		var activatedFpBTCPK *bbn.BIP340PubKey
		for fpBTCPKHex := range inactiveFpsMap {
			stakingValue := uint64(10000000)
			activatedFpBTCPK, _ = bbn.NewBIP340PubKeyFromHex(fpBTCPKHex)
			_, _, _, delMsg, del := h.CreateDelegation(
				r,
				activatedFpBTCPK.MustToBTCPK(),
				changeAddress.EncodeAddress(),
				int64(stakingValue),
				1000,
			)
			h.CreateCovenantSigs(r, covenantSKs, delMsg, del)

			break
		}

		// record voting power table
		babylonHeight += 1
		h.Ctx = datagen.WithCtxHeight(h.Ctx, babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)

		// ensure that the activated finality provider now has entered the active finality provider set
		// i.e., has voting power
		power := h.BTCStakingKeeper.GetVotingPower(h.Ctx, activatedFpBTCPK.MustMarshal(), babylonHeight)
		require.Positive(t, power)
	})
}
