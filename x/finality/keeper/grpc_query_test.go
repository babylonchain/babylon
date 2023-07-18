package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func FuzzListPublicRandomness(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// add a random list of EOTS public randomness
		valBTCPK, err := datagen.GenRandomBIP340PubKey(r)
		require.NoError(t, err)
		startHeight := datagen.RandomInt(r, 100)
		numPubRand := datagen.RandomInt(r, 1000) + 1
		_, prList, err := datagen.GenRandomPubRandList(r, numPubRand)
		require.NoError(t, err)
		keeper.SetPubRandList(ctx, valBTCPK, startHeight, prList)

		// perform a query to pubrand list and assert consistency
		// NOTE: pagination is already tested in Cosmos SDK so we don't test it here again,
		// instead only ensure it takes effect
		limit := datagen.RandomInt(r, int(numPubRand)-1) + 1
		req := &types.QueryListPublicRandomnessRequest{
			ValBtcPkHex: valBTCPK.ToHexStr(),
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}
		resp, err := keeper.ListPublicRandomness(ctx, req)
		require.NoError(t, err)
		require.Equal(t, int(limit), len(resp.PubRandMap)) // check if pagination takes effect
		for i := startHeight; i < startHeight+limit; i++ {
			expectedPR := prList[i-startHeight]
			actualPR := resp.PubRandMap[i]
			require.Equal(t, expectedPR.MustMarshal(), actualPR.MustMarshal())
		}
	})
}

func FuzzListBlocks(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// index a random list of finalised blocks
		startHeight := datagen.RandomInt(r, 100)
		numIndexedBlocks := datagen.RandomInt(r, 100) + 1
		finalizedIndexedBlocks := map[uint64]*types.IndexedBlock{}
		for i := startHeight; i < startHeight+numIndexedBlocks; i++ {
			ib := &types.IndexedBlock{
				Height:         startHeight + numIndexedBlocks,
				LastCommitHash: datagen.GenRandomByteArray(r, 32),
			}
			// randomly finalise some of them
			if datagen.RandomInt(r, 2) == 1 {
				ib.Finalized = true
				finalizedIndexedBlocks[ib.Height] = ib
			}
			// insert to KVStore
			keeper.SetBlock(ctx, ib)
		}

		// perform a query to pubrand list and assert consistency
		// NOTE: pagination is already tested in Cosmos SDK so we don't test it here again,
		// instead only ensure it takes effect
		limit := datagen.RandomInt(r, int(numIndexedBlocks)-1) + 1
		req := &types.QueryListBlocksRequest{
			Finalized: true,
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}
		resp, err := keeper.ListBlocks(ctx, req)
		require.NoError(t, err)
		require.LessOrEqual(t, len(resp.Blocks), int(limit)) // check if pagination takes effect
		for _, actualIB := range resp.Blocks {
			require.Equal(t, finalizedIndexedBlocks[actualIB.Height].LastCommitHash, actualIB.LastCommitHash)
		}
	})
}
