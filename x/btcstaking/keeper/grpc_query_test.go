package keeper_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzActivatedHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// not activated yet
		_, err := keeper.GetBTCStakingActivatedHeight(ctx)
		require.Error(t, err)

		randomActivatedHeight := datagen.RandomInt(r, 100) + 1
		btcVal, err := datagen.GenRandomBTCValidator(r)
		require.NoError(t, err)
		keeper.SetVotingPower(ctx, btcVal.BtcPk.MustMarshal(), randomActivatedHeight, uint64(10))

		// now it's activated
		resp, err := keeper.ActivatedHeight(ctx, &types.QueryActivatedHeightRequest{})
		require.NoError(t, err)
		require.Equal(t, randomActivatedHeight, resp.Height)
	})
}

func FuzzBTCValidators(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// Generate random btc validators and add them to kv store
		btcValsMap := make(map[string]*types.BTCValidator)
		for i := 0; i < int(datagen.RandomInt(r, 10)+1); i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)

			keeper.SetBTCValidator(ctx, btcVal)
			btcValsMap[btcVal.BtcPk.MarshalHex()] = btcVal
		}
		numOfBTCValsInStore := len(btcValsMap)

		// Test nil request
		resp, err := keeper.BTCValidators(ctx, nil)
		if resp != nil {
			t.Errorf("Nil input led to a non-nil response")
		}
		if err == nil {
			t.Errorf("Nil input led to a nil error")
		}

		// Generate a page request with a limit and a nil key
		limit := datagen.RandomInt(r, numOfBTCValsInStore) + 1
		pagination := constructRequestWithLimit(r, limit)
		// Generate the initial query
		req := types.QueryBTCValidatorsRequest{Pagination: pagination}
		// Construct a mapping from the btc vals found to a boolean value
		// Will be used later to evaluate whether all the btc vals were returned
		btcValsFound := make(map[string]bool, 0)

		for i := uint64(0); i < uint64(numOfBTCValsInStore); i += limit {
			resp, err = keeper.BTCValidators(ctx, &req)
			if err != nil {
				t.Errorf("Valid request led to an error %s", err)
			}
			if resp == nil {
				t.Fatalf("Valid request led to a nil response")
			}

			for _, val := range resp.BtcValidators {
				// Check if the pk exists in the map
				if _, ok := btcValsMap[val.BtcPk.MarshalHex()]; !ok {
					t.Fatalf("rpc returned a val that was not created")
				}
				btcValsFound[val.BtcPk.MarshalHex()] = true
			}

			// Construct the next page request
			pagination = constructRequestWithKeyAndLimit(r, resp.Pagination.NextKey, limit)
			req = types.QueryBTCValidatorsRequest{Pagination: pagination}
		}

		if len(btcValsFound) != len(btcValsMap) {
			t.Errorf("Some vals were missed. Got %d while %d were expected", len(btcValsFound), len(btcValsMap))
		}
	})
}

func FuzzBTCValidator(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// Generate random btc validators and add them to kv store
		btcValsMap := make(map[string]*types.BTCValidator)
		for i := 0; i < int(datagen.RandomInt(r, 10)+1); i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)

			keeper.SetBTCValidator(ctx, btcVal)
			btcValsMap[btcVal.BtcPk.MarshalHex()] = btcVal
		}

		// Test nil request
		resp, err := keeper.BTCValidators(ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)

		for k, v := range btcValsMap {
			// Generate a request with a valid key
			req := types.QueryBTCValidatorRequest{ValBtcPkHex: k}
			resp, err := keeper.BTCValidator(ctx, &req)
			if err != nil {
				t.Errorf("Valid request led to an error %s", err)
			}
			if resp == nil {
				t.Fatalf("Valid request led to a nil response")
			}

			// check keys from map matches those in returned response
			require.Equal(t, v.BtcPk.MarshalHex(), resp.BtcValidator.BtcPk.MarshalHex())
			require.Equal(t, v.BabylonPk, resp.BtcValidator.BabylonPk)
		}

		// check some random non exsisting guy
		btcVal, err := datagen.GenRandomBTCValidator(r)
		require.NoError(t, err)
		req := types.QueryBTCValidatorRequest{ValBtcPkHex: btcVal.BtcPk.MarshalHex()}
		respNonExists, err := keeper.BTCValidator(ctx, &req)
		require.Error(t, err)
		require.Nil(t, respNonExists)
		require.True(t, errors.Is(err, types.ErrBTCValNotFound))
	})
}

func FuzzPendingBTCDelegations(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup keeper and context
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		keeper, ctx := testkeeper.BTCStakingKeeper(t, btclcKeeper, btccKeeper)

		// jury and slashing addr
		jurySK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		// Generate a random number of BTC validators
		numBTCVals := datagen.RandomInt(r, 5) + 1
		btcVals := []*types.BTCValidator{}
		for i := uint64(0); i < numBTCVals; i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)
			btcVals = append(btcVals, btcVal)
		}

		// Generate a random number of BTC delegations under each validator
		startHeight := datagen.RandomInt(r, 100) + 1
		endHeight := datagen.RandomInt(r, 1000) + startHeight + btcctypes.DefaultParams().CheckpointFinalizationTimeout + 1
		numBTCDels := datagen.RandomInt(r, 10) + 1
		pendingBtcDelsMap := make(map[string]*types.BTCDelegation)
		for _, btcVal := range btcVals {
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(r, btcVal.BtcPk, delSK, jurySK, slashingAddr.String(), startHeight, endHeight, 10000)
				require.NoError(t, err)
				if datagen.RandomInt(r, 2) == 1 {
					// remove jury sig in random BTC delegations to make them inactive
					btcDel.JurySig = nil
					pendingBtcDelsMap[btcDel.BtcPk.MarshalHex()] = btcDel
				}
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)

				txHash := btcDel.StakingTx.MustGetTxHashStr()
				btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).Times(1)
				delView, err := keeper.BTCDelegation(ctx, &types.QueryBTCDelegationRequest{
					StakingTxHashHex: txHash,
				})
				require.NoError(t, err)
				require.NotNil(t, delView)
			}
		}

		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).Times(1)
		keeper.IndexBTCHeight(ctx)

		// querying paginated BTC delegations and assert
		// Generate a page request with a limit and a nil key
		if len(pendingBtcDelsMap) == 0 {
			return
		}
		limit := datagen.RandomInt(r, len(pendingBtcDelsMap)) + 1
		pagination := constructRequestWithLimit(r, limit)
		req := &types.QueryBTCDelegationsRequest{
			Status:     types.BTCDelegationStatus_PENDING,
			Pagination: pagination,
		}
		for i := uint64(0); i < numBTCDels; i += limit {
			resp, err := keeper.BTCDelegations(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			for _, btcDel := range resp.BtcDelegations {
				_, ok := pendingBtcDelsMap[btcDel.BtcPk.MarshalHex()]
				require.True(t, ok)
			}
			// Construct the next page request
			pagination.Key = resp.Pagination.NextKey
		}
	})
}

func FuzzUnbondingBTCDelegations(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup keeper and context
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		keeper, ctx := testkeeper.BTCStakingKeeper(t, btclcKeeper, btccKeeper)

		// jury and slashing addr
		jurySK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		// Generate a random number of BTC validators
		numBTCVals := datagen.RandomInt(r, 5) + 1
		btcVals := []*types.BTCValidator{}
		for i := uint64(0); i < numBTCVals; i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)
			btcVals = append(btcVals, btcVal)
		}

		// Generate a random number of BTC delegations under each validator
		startHeight := datagen.RandomInt(r, 100) + 1
		endHeight := datagen.RandomInt(r, 1000) + startHeight + btcctypes.DefaultParams().CheckpointFinalizationTimeout + 1
		numBTCDels := datagen.RandomInt(r, 10) + 1
		unbondingBtcDelsMap := make(map[string]*types.BTCDelegation)
		for _, btcVal := range btcVals {
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(r, btcVal.BtcPk, delSK, jurySK, slashingAddr.String(), startHeight, endHeight, 10000)
				require.NoError(t, err)

				if datagen.RandomInt(r, 2) == 1 {
					// add unbonding object in random BTC delegations to make them ready to receive jury sig
					btcDel.BtcUndelegation = &types.BTCUndelegation{
						// doesn't matter what we put here
						ValidatorUnbondingSig: btcDel.JurySig,
					}

					if datagen.RandomInt(r, 2) == 1 {
						// these BTC delegations are unbonded
						btcDel.BtcUndelegation.JuryUnbondingSig = btcDel.JurySig
						btcDel.BtcUndelegation.JurySlashingSig = btcDel.JurySig
					} else {
						// these BTC delegations are unbonding
						unbondingBtcDelsMap[btcDel.BtcPk.MarshalHex()] = btcDel
					}
				}

				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)
			}
		}

		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).Times(1)
		keeper.IndexBTCHeight(ctx)

		// querying paginated BTC delegations and assert
		// Generate a page request with a limit and a nil key
		if len(unbondingBtcDelsMap) == 0 {
			return
		}
		limit := datagen.RandomInt(r, len(unbondingBtcDelsMap)) + 1
		pagination := constructRequestWithLimit(r, limit)
		req := &types.QueryBTCDelegationsRequest{
			Status:     types.BTCDelegationStatus_UNBONDING,
			Pagination: pagination,
		}
		for i := uint64(0); i < numBTCDels; i += limit {
			resp, err := keeper.BTCDelegations(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			for _, btcDel := range resp.BtcDelegations {
				_, ok := unbondingBtcDelsMap[btcDel.BtcPk.MarshalHex()]
				require.True(t, ok)
			}
			// Construct the next page request
			pagination.Key = resp.Pagination.NextKey
		}
	})
}

func FuzzBTCValidatorVotingPowerAtHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)

		// random BTC validator
		btcVal, err := datagen.GenRandomBTCValidator(r)
		require.NoError(t, err)
		// add this BTC validator
		keeper.SetBTCValidator(ctx, btcVal)
		// set random voting power at random height
		randomHeight := datagen.RandomInt(r, 100) + 1
		randomPower := datagen.RandomInt(r, 100) + 1
		keeper.SetVotingPower(ctx, btcVal.BtcPk.MustMarshal(), randomHeight, randomPower)

		req := &types.QueryBTCValidatorPowerAtHeightRequest{
			ValBtcPkHex: btcVal.BtcPk.MarshalHex(),
			Height:      randomHeight,
		}
		resp, err := keeper.BTCValidatorPowerAtHeight(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomPower, resp.VotingPower)
	})
}

func FuzzBTCValidatorCurrentVotingPower(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)

		// random BTC validator
		btcVal, err := datagen.GenRandomBTCValidator(r)
		require.NoError(t, err)
		// add this BTC validator
		keeper.SetBTCValidator(ctx, btcVal)
		// set random voting power at random height
		randomHeight := datagen.RandomInt(r, 100) + 1
		ctx = ctx.WithBlockHeight(int64(randomHeight))
		randomPower := datagen.RandomInt(r, 100) + 1
		keeper.SetVotingPower(ctx, btcVal.BtcPk.MustMarshal(), randomHeight, randomPower)

		// assert voting power at current height
		req := &types.QueryBTCValidatorCurrentPowerRequest{
			ValBtcPkHex: btcVal.BtcPk.MarshalHex(),
		}
		resp, err := keeper.BTCValidatorCurrentPower(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomHeight, resp.Height)
		require.Equal(t, randomPower, resp.VotingPower)

		// if height increments but voting power hasn't recorded yet, then
		// we need to return the height and voting power at the last height
		ctx = ctx.WithBlockHeight(int64(randomHeight + 1))
		resp, err = keeper.BTCValidatorCurrentPower(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomHeight, resp.Height)
		require.Equal(t, randomPower, resp.VotingPower)

		// but no more
		ctx = ctx.WithBlockHeight(int64(randomHeight + 2))
		resp, err = keeper.BTCValidatorCurrentPower(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomHeight+1, resp.Height)
		require.Equal(t, uint64(0), resp.VotingPower)
	})
}

func FuzzActiveBTCValidatorsAtHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)

		// jury and slashing addr
		jurySK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		// Generate a random batch of validators
		var btcVals []*types.BTCValidator
		numBTCValsWithVotingPower := datagen.RandomInt(r, 10) + 1
		numBTCVals := numBTCValsWithVotingPower + datagen.RandomInt(r, 10)
		for i := uint64(0); i < numBTCVals; i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			keeper.SetBTCValidator(ctx, btcVal)
			btcVals = append(btcVals, btcVal)
		}

		// For numBTCValsWithVotingPower validators, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		babylonHeight := datagen.RandomInt(r, 10) + 1
		btcValsWithVotingPowerMap := make(map[string]*types.BTCValidator)
		for i := uint64(0); i < numBTCValsWithVotingPower; i++ {
			valBTCPK := btcVals[i].BtcPk
			btcValsWithVotingPowerMap[valBTCPK.MarshalHex()] = btcVals[i]

			var totalVotingPower uint64
			for j := uint64(0); j < numBTCDels; j++ {
				delSK, _, err := datagen.GenRandomBTCKeyPair(r)
				require.NoError(t, err)
				btcDel, err := datagen.GenRandomBTCDelegation(r, valBTCPK, delSK, jurySK, slashingAddr.String(), 1, 1000, 10000) // timelock period: 1-1000
				require.NoError(t, err)
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)
				totalVotingPower += btcDel.TotalSat
			}

			keeper.SetVotingPower(ctx, valBTCPK.MustMarshal(), babylonHeight, totalVotingPower)
		}

		// Test nil request
		resp, err := keeper.ActiveBTCValidatorsAtHeight(ctx, nil)
		if resp != nil {
			t.Errorf("Nil input led to a non-nil response")
		}
		if err == nil {
			t.Errorf("Nil input led to a nil error")
		}

		// Generate a page request with a limit and a nil key
		limit := datagen.RandomInt(r, int(numBTCValsWithVotingPower)) + 1
		pagination := constructRequestWithLimit(r, limit)
		// Generate the initial query
		req := types.QueryActiveBTCValidatorsAtHeightRequest{Height: babylonHeight, Pagination: pagination}
		// Construct a mapping from the btc vals found to a boolean value
		// Will be used later to evaluate whether all the btc vals were returned
		btcValsFound := make(map[string]bool, 0)

		for i := uint64(0); i < numBTCValsWithVotingPower; i += limit {
			resp, err = keeper.ActiveBTCValidatorsAtHeight(ctx, &req)
			if err != nil {
				t.Errorf("Valid request led to an error %s", err)
			}
			if resp == nil {
				t.Fatalf("Valid request led to a nil response")
			}

			for _, val := range resp.BtcValidators {
				// Check if the pk exists in the map
				if _, ok := btcValsWithVotingPowerMap[val.BtcPk.MarshalHex()]; !ok {
					t.Fatalf("rpc returned a val that was not created")
				}
				btcValsFound[val.BtcPk.MarshalHex()] = true
			}

			// Construct the next page request
			pagination = constructRequestWithKeyAndLimit(r, resp.Pagination.NextKey, limit)
			req = types.QueryActiveBTCValidatorsAtHeightRequest{Height: babylonHeight, Pagination: pagination}
		}

		if len(btcValsFound) != len(btcValsWithVotingPowerMap) {
			t.Errorf("Some vals were missed. Got %d while %d were expected", len(btcValsFound), len(btcValsWithVotingPowerMap))
		}
	})
}

func FuzzBTCValidatorDelegations(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup keeper and context
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		keeper, ctx := testkeeper.BTCStakingKeeper(t, btclcKeeper, btccKeeper)

		// jury and slashing addr
		jurySK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		// Generate a btc validator
		btcVal, err := datagen.GenRandomBTCValidator(r)
		require.NoError(t, err)
		keeper.SetBTCValidator(ctx, btcVal)

		startHeight := datagen.RandomInt(r, 100) + 1
		endHeight := datagen.RandomInt(r, 1000) + startHeight + btcctypes.DefaultParams().CheckpointFinalizationTimeout + 1
		// Generate a random number of BTC delegations under this validator
		numBTCDels := datagen.RandomInt(r, 10) + 1
		expectedBtcDelsMap := make(map[string]*types.BTCDelegation)
		for j := uint64(0); j < numBTCDels; j++ {
			delSK, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcDel, err := datagen.GenRandomBTCDelegation(r, btcVal.BtcPk, delSK, jurySK, slashingAddr.String(), startHeight, endHeight, 10000)
			require.NoError(t, err)
			expectedBtcDelsMap[btcDel.BtcPk.MarshalHex()] = btcDel
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)
		}

		// Test nil request
		resp, err := keeper.BTCValidatorDelegations(ctx, nil)
		require.Nil(t, resp)
		require.Error(t, err)

		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = ctx.WithBlockHeight(int64(babylonHeight))
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).Times(1)
		keeper.IndexBTCHeight(ctx)

		// Generate a page request with a limit and a nil key
		// query a page of BTC delegations and assert consistency
		limit := datagen.RandomInt(r, len(expectedBtcDelsMap)) + 1
		pagination := constructRequestWithLimit(r, limit)
		// Generate the initial query
		req := types.QueryBTCValidatorDelegationsRequest{
			ValBtcPkHex: btcVal.BtcPk.MarshalHex(),
			Pagination:  pagination,
		}
		// Construct a mapping from the btc vals found to a boolean value
		// Will be used later to evaluate whether all the btc vals were returned
		btcDelsFound := make(map[string]bool, 0)

		for i := uint64(0); i < numBTCDels; i += limit {
			resp, err = keeper.BTCValidatorDelegations(ctx, &req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			for _, btcDels := range resp.BtcDelegatorDelegations {
				require.Len(t, btcDels.Dels, 1)
				btcDel := btcDels.Dels[0]
				require.Equal(t, btcVal.BtcPk, btcDel.ValBtcPk)
				// Check if the pk exists in the map
				_, ok := expectedBtcDelsMap[btcDel.BtcPk.MarshalHex()]
				require.True(t, ok)
				btcDelsFound[btcDel.BtcPk.MarshalHex()] = true
			}
			// Construct the next page request
			pagination = constructRequestWithKeyAndLimit(r, resp.Pagination.NextKey, limit)
			req = types.QueryBTCValidatorDelegationsRequest{
				ValBtcPkHex: btcVal.BtcPk.MarshalHex(),
				Pagination:  pagination,
			}
		}
		require.Equal(t, len(btcDelsFound), len(expectedBtcDelsMap))

	})
}

// Constructors for PageRequest objects
func constructRequestWithKeyAndLimit(r *rand.Rand, key []byte, limit uint64) *query.PageRequest {
	// If limit is 0, set one randomly
	if limit == 0 {
		limit = uint64(r.Int63() + 1) // Use Int63 instead of Uint64 to avoid overflows
	}
	return &query.PageRequest{
		Key:        key,
		Offset:     0, // only offset or key is set
		Limit:      limit,
		CountTotal: false, // only used when offset is used
		Reverse:    false,
	}
}

func constructRequestWithLimit(r *rand.Rand, limit uint64) *query.PageRequest {
	return constructRequestWithKeyAndLimit(r, nil, limit)
}
