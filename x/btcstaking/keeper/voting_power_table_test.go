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
		fps := []*types.FinalityProvider{}
		numFpsWithVotingPower := datagen.RandomInt(r, 10) + 2
		numFps := numFpsWithVotingPower + datagen.RandomInt(r, 10)
		for i := uint64(0); i < numFps; i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			keeper.SetFinalityProvider(ctx, fp)
			fps = append(fps, fp)
		}

		// for the first numFpsWithVotingPower finality providers, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		stakingValue := datagen.RandomInt(r, 100000) + 100000
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(
					r,
					t,
					[]bbn.BIP340PubKey{*fps[i].BtcPk},
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

		/*
			Case 1: BTC height is 0 that is smaller than start height 1.
			No finality privater will have voting power
		*/
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 0}).Times(1)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		for i := uint64(0); i < numFps; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		/*
			Case 2: move to 1st BTC block, then assert the first numFpsWithVotingPower finality providers have voting power
		*/
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			require.Equal(t, numBTCDels*stakingValue, power)
		}
		for i := numFpsWithVotingPower; i < numFps; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// also, get voting power table and assert consistency
		powerTable := keeper.GetVotingPowerTable(ctx, babylonHeight)
		require.NotNil(t, powerTable)
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			require.Equal(t, powerTable[fps[i].BtcPk.MarshalHex()], power)
		}
		// the activation height should be the current Babylon height as well
		activatedHeight, err := keeper.GetBTCStakingActivatedHeight(ctx)
		require.NoError(t, err)
		require.Equal(t, babylonHeight, activatedHeight)

		/*
			Case 3: slash a random finality provider and move on
			then assert the slashed finality provider does not have voting power
		*/
		// slash a random finality provider
		slashedIdx := datagen.RandomInt(r, int(numFpsWithVotingPower))
		slashedFp := fps[slashedIdx]
		// This will be called to get the slashed height
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		err = keeper.SlashFinalityProvider(ctx, slashedFp.BtcPk.MustMarshal())
		require.NoError(t, err)
		// move to later Babylon height and 2nd BTC height
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 2}).Times(1)
		// index height and record power table
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		// check if the slashed finality provider's voting power becomes zero
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			if i == slashedIdx {
				require.Zero(t, power)
			} else {
				require.Equal(t, numBTCDels*stakingValue, power)
			}
		}
		for i := numFpsWithVotingPower; i < numFps; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// also, get voting power table and assert consistency
		powerTable = keeper.GetVotingPowerTable(ctx, babylonHeight)
		require.NotNil(t, powerTable)
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *fps[i].BtcPk, babylonHeight)
			if i == slashedIdx {
				require.Zero(t, power)
			}
			require.Equal(t, powerTable[fps[i].BtcPk.MarshalHex()], power)
		}

		/*
			Case 4: move to 999th BTC block, then assert none of finality providers has voting power (since end height - w < BTC height)
		*/
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 999}).Times(1)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		for _, fp := range fps {
			power := keeper.GetVotingPower(ctx, *fp.BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// the activation height should be same as before
		activatedHeight2, err := keeper.GetBTCStakingActivatedHeight(ctx)
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

		// generate a random batch of finality providers, each with a BTC delegation with random power
		fpsWithMeta := []*types.FinalityProviderDistInfo{}
		numFps := datagen.RandomInt(r, 300) + 1
		for i := uint64(0); i < numFps; i++ {
			// generate finality provider
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			keeper.SetFinalityProvider(ctx, fp)

			// delegate to this finality provider
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			fpBTCPK := fp.BtcPk
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(
				r,
				t,
				[]bbn.BIP340PubKey{*fpBTCPK},
				delSK,
				covenantSKs,
				covenantQuorum,
				slashingAddress.EncodeAddress(),
				1, 1000, stakingValue, // timelock period: 1-1000
				slashingRate,
				slashingChangeLockTime,
			)
			require.NoError(t, err)
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)

			// record voting power
			fpsWithMeta = append(fpsWithMeta, &types.FinalityProviderDistInfo{
				BtcPk:            fp.BtcPk,
				TotalVotingPower: stakingValue,
			})
		}

		maxActiveFpsParam := keeper.GetParams(ctx).MaxActiveFinalityProviders
		// get a map of expected active finality providers
		expectedActiveFps := types.FilterTopNFinalityProviders(fpsWithMeta, maxActiveFpsParam)
		expectedActiveFpsMap := map[string]uint64{}
		for _, fp := range expectedActiveFps {
			expectedActiveFpsMap[fp.BtcPk.MarshalHex()] = fp.TotalVotingPower
		}

		// record voting power table
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		//  only finality providers in expectedActiveFpsMap have voting power
		for _, fp := range fpsWithMeta {
			power := keeper.GetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonHeight)
			if expectedPower, ok := expectedActiveFpsMap[fp.BtcPk.MarshalHex()]; ok {
				require.Equal(t, expectedPower, power)
			} else {
				require.Equal(t, uint64(0), power)
			}
		}

		// also, get voting power table and assert there is
		// min(len(expectedActiveFps), MaxActiveFinalityProviders) active finality providers
		powerTable := keeper.GetVotingPowerTable(ctx, babylonHeight)
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

		// generate a random batch of finality providers, each with a BTC delegation with random power
		fpsWithMeta := []*types.FinalityProviderWithMeta{}
		numFps := uint64(200) // there has to be more than `maxActiveFinalityProviders` finality providers
		for i := uint64(0); i < numFps; i++ {
			// generate finality provider
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			keeper.SetFinalityProvider(ctx, fp)

			// delegate to this finality provider
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			fpBTCPK := fp.BtcPk
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(
				r,
				t,
				[]bbn.BIP340PubKey{*fpBTCPK},
				delSK,
				covenantSKs,
				covenantQuorum,
				slashingAddress.EncodeAddress(),
				1, 1000, stakingValue, // timelock period: 1-1000
				slashingRate,
				slashingChangeLockTime,
			)
			require.NoError(t, err)
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)

			// record voting power
			fpsWithMeta = append(fpsWithMeta, &types.FinalityProviderWithMeta{
				BtcPk:       fp.BtcPk,
				VotingPower: stakingValue,
			})
		}

		// record voting power table
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		// get maps of active/inactive finality providers
		activeFpsMap := map[string]uint64{}
		inactiveFpsMap := map[string]uint64{}
		for _, fp := range fpsWithMeta {
			power := keeper.GetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonHeight)
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
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(
				r,
				t,
				[]bbn.BIP340PubKey{*activatedFpBTCPK},
				delSK,
				covenantSKs,
				covenantQuorum,
				slashingAddress.EncodeAddress(),
				1, 1000, stakingValue, // timelock period: 1-1000
				slashingRate,
				slashingChangeLockTime,
			)
			require.NoError(t, err)
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)

			break
		}

		// record voting power table
		babylonHeight += 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		err = keeper.BeginBlocker(ctx)
		require.NoError(t, err)

		// ensure that the activated finality provider now has entered the active finality provider set
		// i.e., has voting power
		power := keeper.GetVotingPower(ctx, activatedFpBTCPK.MustMarshal(), babylonHeight)
		require.Positive(t, power)
	})
}
