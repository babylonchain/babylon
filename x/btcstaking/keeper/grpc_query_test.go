package keeper_test

import (
	"errors"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
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
		fp, err := datagen.GenRandomFinalityProvider(r)
		require.NoError(t, err)
		keeper.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), randomActivatedHeight, uint64(10))

		// now it's activated
		resp, err := keeper.ActivatedHeight(ctx, &types.QueryActivatedHeightRequest{})
		require.NoError(t, err)
		require.Equal(t, randomActivatedHeight, resp.Height)
	})
}

func FuzzFinalityProviders(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// Generate random finality providers and add them to kv store
		fpsMap := make(map[string]*types.FinalityProvider)
		for i := 0; i < int(datagen.RandomInt(r, 10)+1); i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)

			keeper.SetFinalityProvider(ctx, fp)
			fpsMap[fp.BtcPk.MarshalHex()] = fp
		}
		numOfFpsInStore := len(fpsMap)

		// Test nil request
		resp, err := keeper.FinalityProviders(ctx, nil)
		if resp != nil {
			t.Errorf("Nil input led to a non-nil response")
		}
		if err == nil {
			t.Errorf("Nil input led to a nil error")
		}

		// Generate a page request with a limit and a nil key
		limit := datagen.RandomInt(r, numOfFpsInStore) + 1
		pagination := constructRequestWithLimit(r, limit)
		// Generate the initial query
		req := types.QueryFinalityProvidersRequest{Pagination: pagination}
		// Construct a mapping from the finality providers found to a boolean value
		// Will be used later to evaluate whether all the finality providers were returned
		fpsFound := make(map[string]bool, 0)

		for i := uint64(0); i < uint64(numOfFpsInStore); i += limit {
			resp, err = keeper.FinalityProviders(ctx, &req)
			if err != nil {
				t.Errorf("Valid request led to an error %s", err)
			}
			if resp == nil {
				t.Fatalf("Valid request led to a nil response")
			}

			for _, fp := range resp.FinalityProviders {
				// Check if the pk exists in the map
				if _, ok := fpsMap[fp.BtcPk.MarshalHex()]; !ok {
					t.Fatalf("rpc returned a finality provider that was not created")
				}
				fpsFound[fp.BtcPk.MarshalHex()] = true
			}

			// Construct the next page request
			pagination = constructRequestWithKeyAndLimit(r, resp.Pagination.NextKey, limit)
			req = types.QueryFinalityProvidersRequest{Pagination: pagination}
		}

		if len(fpsFound) != len(fpsMap) {
			t.Errorf("Some finality providers were missed. Got %d while %d were expected", len(fpsFound), len(fpsMap))
		}
	})
}

func FuzzFinalityProvider(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// Generate random finality providers and add them to kv store
		fpsMap := make(map[string]*types.FinalityProvider)
		for i := 0; i < int(datagen.RandomInt(r, 10)+1); i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)

			keeper.SetFinalityProvider(ctx, fp)
			fpsMap[fp.BtcPk.MarshalHex()] = fp
		}

		// Test nil request
		resp, err := keeper.FinalityProvider(ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)

		for k, v := range fpsMap {
			// Generate a request with a valid key
			req := types.QueryFinalityProviderRequest{FpBtcPkHex: k}
			resp, err := keeper.FinalityProvider(ctx, &req)
			if err != nil {
				t.Errorf("Valid request led to an error %s", err)
			}
			if resp == nil {
				t.Fatalf("Valid request led to a nil response")
			}

			// check keys from map matches those in returned response
			require.Equal(t, v.BtcPk.MarshalHex(), resp.FinalityProvider.BtcPk.MarshalHex())
			require.Equal(t, v.BabylonPk, resp.FinalityProvider.BabylonPk)
		}

		// check some random non-existing guy
		fp, err := datagen.GenRandomFinalityProvider(r)
		require.NoError(t, err)
		req := types.QueryFinalityProviderRequest{FpBtcPkHex: fp.BtcPk.MarshalHex()}
		respNonExists, err := keeper.FinalityProvider(ctx, &req)
		require.Error(t, err)
		require.Nil(t, respNonExists)
		require.True(t, errors.Is(err, types.ErrFpNotFound))
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

		// Generate a random number of finality providers
		numFps := datagen.RandomInt(r, 5) + 1
		fps := []*types.FinalityProvider{}
		for i := uint64(0); i < numFps; i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			keeper.SetFinalityProvider(ctx, fp)
			fps = append(fps, fp)
		}

		// Generate a random number of BTC delegations under each finality provider
		startHeight := datagen.RandomInt(r, 100) + 1
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).AnyTimes()

		endHeight := datagen.RandomInt(r, 1000) + startHeight + btcctypes.DefaultParams().CheckpointFinalizationTimeout + 1
		numBTCDels := datagen.RandomInt(r, 10) + 1
		pendingBtcDelsMap := make(map[string]*types.BTCDelegation)
		for _, fp := range fps {
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
					startHeight, endHeight, 10000,
					slashingRate,
					slashingChangeLockTime,
				)
				require.NoError(t, err)
				if datagen.RandomInt(r, 2) == 1 {
					// remove covenant sig in random BTC delegations to make them inactive
					btcDel.CovenantSigs = nil
					pendingBtcDelsMap[btcDel.BtcPk.MarshalHex()] = btcDel
				}
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)

				txHash := btcDel.MustGetStakingTxHash().String()
				delView, err := keeper.BTCDelegation(ctx, &types.QueryBTCDelegationRequest{
					StakingTxHashHex: txHash,
				})
				require.NoError(t, err)
				require.NotNil(t, delView)
			}
		}

		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)

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

func FuzzFinalityProviderPowerAtHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)

		// random finality provider
		fp, err := datagen.GenRandomFinalityProvider(r)
		require.NoError(t, err)
		// add this finality provider
		keeper.SetFinalityProvider(ctx, fp)
		// set random voting power at random height
		randomHeight := datagen.RandomInt(r, 100) + 1
		randomPower := datagen.RandomInt(r, 100) + 1
		keeper.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), randomHeight, randomPower)

		req := &types.QueryFinalityProviderPowerAtHeightRequest{
			FpBtcPkHex: fp.BtcPk.MarshalHex(),
			Height:     randomHeight,
		}
		resp, err := keeper.FinalityProviderPowerAtHeight(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomPower, resp.VotingPower)
	})
}

func FuzzFinalityProviderCurrentVotingPower(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.BTCStakingKeeper(t, nil, nil)

		// random finality provider
		fp, err := datagen.GenRandomFinalityProvider(r)
		require.NoError(t, err)
		// add this finality provider
		keeper.SetFinalityProvider(ctx, fp)
		// set random voting power at random height
		randomHeight := datagen.RandomInt(r, 100) + 1
		ctx = datagen.WithCtxHeight(ctx, randomHeight)
		randomPower := datagen.RandomInt(r, 100) + 1
		keeper.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), randomHeight, randomPower)

		// assert voting power at current height
		req := &types.QueryFinalityProviderCurrentPowerRequest{
			FpBtcPkHex: fp.BtcPk.MarshalHex(),
		}
		resp, err := keeper.FinalityProviderCurrentPower(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomHeight, resp.Height)
		require.Equal(t, randomPower, resp.VotingPower)

		// if height increments but voting power hasn't recorded yet, then
		// we need to return the height and voting power at the last height
		ctx = datagen.WithCtxHeight(ctx, randomHeight+1)
		resp, err = keeper.FinalityProviderCurrentPower(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomHeight, resp.Height)
		require.Equal(t, randomPower, resp.VotingPower)

		// test the case when the finality provider has 0 voting power
		ctx = datagen.WithCtxHeight(ctx, randomHeight+2)
		keeper.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), randomHeight+2, 0)
		resp, err = keeper.FinalityProviderCurrentPower(ctx, req)
		require.NoError(t, err)
		require.Equal(t, randomHeight+2, resp.Height)
		require.Equal(t, uint64(0), resp.VotingPower)
	})
}

func FuzzActiveFinalityProvidersAtHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup keeper and context
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 10}).AnyTimes()
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		keeper, ctx := testkeeper.BTCStakingKeeper(t, btclcKeeper, btccKeeper)

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

		// Generate a random batch of finality providers
		var fps []*types.FinalityProvider
		numFpsWithVotingPower := datagen.RandomInt(r, 10) + 1
		numFps := numFpsWithVotingPower + datagen.RandomInt(r, 10)
		for i := uint64(0); i < numFps; i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			keeper.SetFinalityProvider(ctx, fp)
			fps = append(fps, fp)
		}

		// For numFpsWithVotingPower finality providers, generate a random number of BTC delegations
		numBTCDels := datagen.RandomInt(r, 10) + 1
		babylonHeight := datagen.RandomInt(r, 10) + 1
		fpsWithVotingPowerMap := make(map[string]*types.FinalityProvider)
		for i := uint64(0); i < numFpsWithVotingPower; i++ {
			fpBTCPK := fps[i].BtcPk
			fpsWithVotingPowerMap[fpBTCPK.MarshalHex()] = fps[i]

			var totalVotingPower uint64
			for j := uint64(0); j < numBTCDels; j++ {
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
					1, 1000, 10000,
					slashingRate,
					slashingChangeLockTime,
				)
				require.NoError(t, err)
				err = keeper.AddBTCDelegation(ctx, btcDel)
				require.NoError(t, err)
				totalVotingPower += btcDel.TotalSat
			}

			keeper.SetVotingPower(ctx, fpBTCPK.MustMarshal(), babylonHeight, totalVotingPower)
		}

		// Test nil request
		resp, err := keeper.ActiveFinalityProvidersAtHeight(ctx, nil)
		if resp != nil {
			t.Errorf("Nil input led to a non-nil response")
		}
		if err == nil {
			t.Errorf("Nil input led to a nil error")
		}

		// Generate a page request with a limit and a nil key
		limit := datagen.RandomInt(r, int(numFpsWithVotingPower)) + 1
		pagination := constructRequestWithLimit(r, limit)
		// Generate the initial query
		req := types.QueryActiveFinalityProvidersAtHeightRequest{Height: babylonHeight, Pagination: pagination}
		// Construct a mapping from the finality providers found to a boolean value
		// Will be used later to evaluate whether all the finality providers were returned
		fpsFound := make(map[string]bool, 0)

		for i := uint64(0); i < numFpsWithVotingPower; i += limit {
			resp, err = keeper.ActiveFinalityProvidersAtHeight(ctx, &req)
			if err != nil {
				t.Errorf("Valid request led to an error %s", err)
			}
			if resp == nil {
				t.Fatalf("Valid request led to a nil response")
			}

			for _, fp := range resp.FinalityProviders {
				// Check if the pk exists in the map
				if _, ok := fpsWithVotingPowerMap[fp.BtcPk.MarshalHex()]; !ok {
					t.Fatalf("rpc returned a finality provider that was not created")
				}
				fpsFound[fp.BtcPk.MarshalHex()] = true
			}

			// Construct the next page request
			pagination = constructRequestWithKeyAndLimit(r, resp.Pagination.NextKey, limit)
			req = types.QueryActiveFinalityProvidersAtHeightRequest{Height: babylonHeight, Pagination: pagination}
		}

		if len(fpsFound) != len(fpsWithVotingPowerMap) {
			t.Errorf("Some finality providers were missed. Got %d while %d were expected", len(fpsFound), len(fpsWithVotingPowerMap))
		}
	})
}

func FuzzFinalityProviderDelegations(f *testing.F) {
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

		// Generate a finality provider
		fp, err := datagen.GenRandomFinalityProvider(r)
		require.NoError(t, err)
		keeper.SetFinalityProvider(ctx, fp)

		startHeight := datagen.RandomInt(r, 100) + 1
		endHeight := datagen.RandomInt(r, 1000) + startHeight + btcctypes.DefaultParams().CheckpointFinalizationTimeout + 1
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).AnyTimes()
		// Generate a random number of BTC delegations under this finality provider
		numBTCDels := datagen.RandomInt(r, 10) + 1
		expectedBtcDelsMap := make(map[string]*types.BTCDelegation)
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
				startHeight, endHeight, 10000,
				slashingRate,
				slashingChangeLockTime,
			)
			require.NoError(t, err)
			expectedBtcDelsMap[btcDel.BtcPk.MarshalHex()] = btcDel
			err = keeper.AddBTCDelegation(ctx, btcDel)
			require.NoError(t, err)
		}

		// Test nil request
		resp, err := keeper.FinalityProviderDelegations(ctx, nil)
		require.Nil(t, resp)
		require.Error(t, err)

		babylonHeight := datagen.RandomInt(r, 10) + 1
		ctx = datagen.WithCtxHeight(ctx, babylonHeight)
		keeper.IndexBTCHeight(ctx)

		// Generate a page request with a limit and a nil key
		// query a page of BTC delegations and assert consistency
		limit := datagen.RandomInt(r, len(expectedBtcDelsMap)) + 1

		// FinalityProviderDelegations loads status, which calls GetTipInfo
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: startHeight}).AnyTimes()

		keeper.IndexBTCHeight(ctx)

		pagination := constructRequestWithLimit(r, limit)
		// Generate the initial query
		req := types.QueryFinalityProviderDelegationsRequest{
			FpBtcPkHex: fp.BtcPk.MarshalHex(),
			Pagination: pagination,
		}
		// Construct a mapping from the finality providers found to a boolean value
		// Will be used later to evaluate whether all the finality providers were returned
		btcDelsFound := make(map[string]bool, 0)

		for i := uint64(0); i < numBTCDels; i += limit {
			resp, err = keeper.FinalityProviderDelegations(ctx, &req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			for _, btcDels := range resp.BtcDelegatorDelegations {
				require.Len(t, btcDels.Dels, 1)
				btcDel := btcDels.Dels[0]
				require.Equal(t, fp.BtcPk, &btcDel.FpBtcPkList[0])
				// Check if the pk exists in the map
				_, ok := expectedBtcDelsMap[btcDel.BtcPk.MarshalHex()]
				require.True(t, ok)
				btcDelsFound[btcDel.BtcPk.MarshalHex()] = true
			}
			// Construct the next page request
			pagination = constructRequestWithKeyAndLimit(r, resp.Pagination.NextKey, limit)
			req = types.QueryFinalityProviderDelegationsRequest{
				FpBtcPkHex: fp.BtcPk.MarshalHex(),
				Pagination: pagination,
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
