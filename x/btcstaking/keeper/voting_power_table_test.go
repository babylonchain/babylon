package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		covenantSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		changeAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdk.NewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// generate a random batch of validators
		btcVals := []*types.BTCValidator{}
		numBTCValsWithVotingPower := datagen.RandomInt(r, 10) + 2
		numBTCVals := numBTCValsWithVotingPower + datagen.RandomInt(r, 10)
		for i := uint64(0); i < numBTCVals; i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)
			btcVals = append(btcVals, btcVal)
		}

		// for the first numBTCValsWithVotingPower validators, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		stakingValue := datagen.RandomInt(r, 100000) + 100000
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(
					r,
					[]bbn.BIP340PubKey{*btcVals[i].BtcPk},
					delSK,
					[]*btcec.PrivateKey{covenantSK},
					slashingAddress.String(), changeAddress.String(),
					1, 1000, stakingValue,
					slashingRate,
				)
				require.NoError(t, err)
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)
			}
		}

		/*
			Case 1: assert none of validators has voting power (since BTC height is 0)
		*/
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
		_, err = keeper.GetBTCStakingActivatedHeight(ctx)
		require.Error(t, err)

		/*
			Case 2: move to 1st BTC block, then assert the first numBTCValsWithVotingPower validators have voting power
		*/
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			require.Equal(t, uint64(numBTCDels)*stakingValue, power)
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
			require.Equal(t, powerTable[btcVals[i].BtcPk.MarshalHex()], power)
		}
		// the activation height should be the current Babylon height as well
		activatedHeight, err := keeper.GetBTCStakingActivatedHeight(ctx)
		require.NoError(t, err)
		require.Equal(t, babylonHeight, activatedHeight)

		/*
			Case 3: slash a random BTC validator and move on
			then assert the slashed BTC validator does not have voting power
		*/
		// slash a random BTC validator
		slashedIdx := datagen.RandomInt(r, int(numBTCValsWithVotingPower))
		slashedVal := btcVals[slashedIdx]
		// This will be called to get the slashed height
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		err = keeper.SlashBTCValidator(ctx, slashedVal.BtcPk.MustMarshal())
		require.NoError(t, err)
		// move to later Babylon height and 2nd BTC height
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 2}).Times(1)
		// index height and record power table
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)
		// check if the slashed BTC validator's voting power becomes zero
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			if i == slashedIdx {
				require.Zero(t, power)
			} else {
				require.Equal(t, uint64(numBTCDels)*stakingValue, power)
			}
		}
		for i := numBTCValsWithVotingPower; i < numBTCVals; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// also, get voting power table and assert consistency
		powerTable = keeper.GetVotingPowerTable(ctx, babylonHeight)
		require.NotNil(t, powerTable)
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			power := keeper.GetVotingPower(ctx, *btcVals[i].BtcPk, babylonHeight)
			if i == slashedIdx {
				require.Zero(t, power)
			}
			require.Equal(t, powerTable[btcVals[i].BtcPk.MarshalHex()], power)
		}

		/*
			Case 4: move to 999th BTC block, then assert none of validators has voting power (since end height - w < BTC height)
		*/
		babylonHeight += datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 999}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)
		for _, btcVal := range btcVals {
			power := keeper.GetVotingPower(ctx, *btcVal.BtcPk, babylonHeight)
			require.Zero(t, power)
		}

		// the activation height should be same as before
		activatedHeight2, err := keeper.GetBTCStakingActivatedHeight(ctx)
		require.NoError(t, err)
		require.Equal(t, activatedHeight, activatedHeight2)
	})
}

func FuzzVotingPowerTable_ActiveBTCValidators(f *testing.F) {
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
		covenantSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		changeAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdk.NewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// generate a random batch of validators, each with a BTC delegation with random power
		btcValsWithMeta := []*types.BTCValidatorWithMeta{}
		numBTCVals := datagen.RandomInt(r, 300) + 1
		for i := uint64(0); i < numBTCVals; i++ {
			// generate BTC validator
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)

			// delegate to this BTC validator
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			valBTCPK := btcVal.BtcPk
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(
				r,
				[]bbn.BIP340PubKey{*valBTCPK},
				delSK,
				[]*btcec.PrivateKey{covenantSK},
				slashingAddress.String(), changeAddress.String(),
				1, 1000, stakingValue, // timelock period: 1-1000
				slashingRate,
			)
			require.NoError(t, err)
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)

			// record voting power
			btcValsWithMeta = append(btcValsWithMeta, &types.BTCValidatorWithMeta{
				BtcPk:       btcVal.BtcPk,
				VotingPower: stakingValue,
			})
		}

		maxActiveBTCValsParam := keeper.GetParams(ctx).MaxActiveBtcValidators
		// get a map of expected active BTC validators
		expectedActiveBTCVals := types.FilterTopNBTCValidators(btcValsWithMeta, maxActiveBTCValsParam)
		expectedActiveBTCValMap := map[string]uint64{}
		for _, btcVal := range expectedActiveBTCVals {
			expectedActiveBTCValMap[btcVal.BtcPk.MarshalHex()] = btcVal.VotingPower
		}

		// record voting power table
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)

		//  only BTC validators in expectedActiveBTCValMap have voting power
		for _, btcVal := range btcValsWithMeta {
			power := keeper.GetVotingPower(ctx, btcVal.BtcPk.MustMarshal(), babylonHeight)
			if expectedPower, ok := expectedActiveBTCValMap[btcVal.BtcPk.MarshalHex()]; ok {
				require.Equal(t, expectedPower, power)
			} else {
				require.Equal(t, uint64(0), power)
			}
		}

		// also, get voting power table and assert there is
		// min(len(expectedActiveBTCVals), MaxActiveBtcValidators) active BTC validators
		powerTable := keeper.GetVotingPowerTable(ctx, babylonHeight)
		expectedNumActiveBTCVals := len(expectedActiveBTCValMap)
		if expectedNumActiveBTCVals > int(maxActiveBTCValsParam) {
			expectedNumActiveBTCVals = int(maxActiveBTCValsParam)
		}
		require.Len(t, powerTable, expectedNumActiveBTCVals)
		// assert consistency of voting power
		for pkHex, expectedPower := range expectedActiveBTCValMap {
			require.Equal(t, powerTable[pkHex], expectedPower)
		}
	})
}

func FuzzVotingPowerTable_ActiveBTCValidatorRotation(f *testing.F) {
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
		covenantSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		changeAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdk.NewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// generate a random batch of validators, each with a BTC delegation with random power
		btcValsWithMeta := []*types.BTCValidatorWithMeta{}
		numBTCVals := uint64(200) // there has to be more than `maxActiveBtcValidators` validators
		for i := uint64(0); i < numBTCVals; i++ {
			// generate BTC validator
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)

			// delegate to this BTC validator
			stakingValue := datagen.RandomInt(r, 100000) + 100000
			valBTCPK := btcVal.BtcPk
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(
				r,
				[]bbn.BIP340PubKey{*valBTCPK},
				delSK,
				[]*btcec.PrivateKey{covenantSK},
				slashingAddress.String(), changeAddress.String(),
				1, 1000, stakingValue, // timelock period: 1-1000
				slashingRate,
			)
			require.NoError(t, err)
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)

			// record voting power
			btcValsWithMeta = append(btcValsWithMeta, &types.BTCValidatorWithMeta{
				BtcPk:       btcVal.BtcPk,
				VotingPower: stakingValue,
			})
		}

		// record voting power table
		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)

		// get maps of active/inactive BTC validators
		activeBTCValMap := map[string]uint64{}
		inactiveBTCValMap := map[string]uint64{}
		for _, btcVal := range btcValsWithMeta {
			power := keeper.GetVotingPower(ctx, btcVal.BtcPk.MustMarshal(), babylonHeight)
			if power > 0 {
				activeBTCValMap[btcVal.BtcPk.MarshalHex()] = power
			} else {
				inactiveBTCValMap[btcVal.BtcPk.MarshalHex()] = power
			}
		}

		// delegate a huge amount of tokens to one of the inactive BTC validator
		var activatedValBTCPK *bbn.BIP340PubKey
		for valBTCPKHex := range inactiveBTCValMap {
			stakingValue := uint64(10000000)
			activatedValBTCPK, _ = bbn.NewBIP340PubKeyFromHex(valBTCPKHex)
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(
				r,
				[]bbn.BIP340PubKey{*activatedValBTCPK},
				delSK,
				[]*btcec.PrivateKey{covenantSK},
				slashingAddress.String(), changeAddress.String(),
				1, 1000, stakingValue, // timelock period: 1-1000
				slashingRate,
			)
			require.NoError(t, err)
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)

			break
		}

		// record voting power table
		babylonHeight += 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 1}).Times(1)
		keeper.IndexBTCHeight(ctx)
		keeper.RecordVotingPowerTable(ctx)

		// ensure that the activated BTC validator now has entered the active validator set
		// i.e., has voting power
		power := keeper.GetVotingPower(ctx, activatedValBTCPK.MustMarshal(), babylonHeight)
		require.Positive(t, power)
	})
}
