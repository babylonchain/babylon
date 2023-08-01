package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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
			ValBtcPkHex: valBTCPK.MarshalHex(),
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
		finalizedIndexedBlocks := make(map[uint64]*types.IndexedBlock)
		nonFinalizedIndexedBlocks := make(map[uint64]*types.IndexedBlock)
		indexedBlocks := make(map[uint64]*types.IndexedBlock)
		for i := startHeight; i < startHeight+numIndexedBlocks; i++ {
			ib := &types.IndexedBlock{
				Height:         i,
				LastCommitHash: datagen.GenRandomByteArray(r, 32),
			}
			// randomly finalise some of them
			if datagen.RandomInt(r, 2) == 1 {
				ib.Finalized = true
				finalizedIndexedBlocks[ib.Height] = ib
			} else {
				nonFinalizedIndexedBlocks[ib.Height] = ib
			}
			indexedBlocks[ib.Height] = ib
			// insert to KVStore
			keeper.SetBlock(ctx, ib)
		}

		// perform a query to fetch finalized blocks and assert consistency
		// NOTE: pagination is already tested in Cosmos SDK so we don't test it here again,
		// instead only ensure it takes effect
		if len(finalizedIndexedBlocks) != 0 {
			limit := datagen.RandomInt(r, len(finalizedIndexedBlocks)) + 1
			req := &types.QueryListBlocksRequest{
				Status: types.QueriedBlockStatus_FINALIZED,
				Pagination: &query.PageRequest{
					CountTotal: true,
					Limit:      limit,
				},
			}
			resp1, err := keeper.ListBlocks(ctx, req)
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp1.Blocks), int(limit)) // check if pagination takes effect
			require.EqualValues(t, resp1.Pagination.Total, len(finalizedIndexedBlocks))
			for _, actualIB := range resp1.Blocks {
				require.Equal(t, finalizedIndexedBlocks[actualIB.Height].LastCommitHash, actualIB.LastCommitHash)
			}
		}

		if len(nonFinalizedIndexedBlocks) != 0 {
			// perform a query to fetch non-finalized blocks and assert consistency
			limit := datagen.RandomInt(r, len(nonFinalizedIndexedBlocks)) + 1
			req := &types.QueryListBlocksRequest{
				Status: types.QueriedBlockStatus_NON_FINALIZED,
				Pagination: &query.PageRequest{
					CountTotal: true,
					Limit:      limit,
				},
			}
			resp2, err := keeper.ListBlocks(ctx, req)
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp2.Blocks), int(limit)) // check if pagination takes effect
			require.EqualValues(t, resp2.Pagination.Total, len(nonFinalizedIndexedBlocks))
			for _, actualIB := range resp2.Blocks {
				require.Equal(t, nonFinalizedIndexedBlocks[actualIB.Height].LastCommitHash, actualIB.LastCommitHash)
			}
		}

		// perform a query to fetch all blocks and assert consistency
		limit := datagen.RandomInt(r, len(indexedBlocks)) + 1
		req := &types.QueryListBlocksRequest{
			Status: types.QueriedBlockStatus_ANY,
			Pagination: &query.PageRequest{
				CountTotal: true,
				Limit:      limit,
			},
		}
		resp3, err := keeper.ListBlocks(ctx, req)
		require.NoError(t, err)
		require.LessOrEqual(t, len(resp3.Blocks), int(limit)) // check if pagination takes effect
		require.EqualValues(t, resp3.Pagination.Total, len(indexedBlocks))
		for _, actualIB := range resp3.Blocks {
			require.Equal(t, indexedBlocks[actualIB.Height].LastCommitHash, actualIB.LastCommitHash)
		}
	})
}

func FuzzVotesAtHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// Add random number of voted validators to the store
		babylonHeight := datagen.RandomInt(r, 10) + 1
		numVotedVals := datagen.RandomInt(r, 10) + 1
		votedValMap := make(map[string]bool, numVotedVals)
		for i := uint64(0); i < numVotedVals; i++ {
			votedValPK, err := datagen.GenRandomBIP340PubKey(r)
			require.NoError(t, err)
			votedSig, err := bbn.NewSchnorrEOTSSig(datagen.GenRandomByteArray(r, 32))
			require.NoError(t, err)
			keeper.SetSig(ctx, babylonHeight, votedValPK, votedSig)

			votedValMap[votedValPK.MarshalHex()] = true
		}

		resp, err := keeper.VotesAtHeight(ctx, &types.QueryVotesAtHeightRequest{
			Height: babylonHeight,
		})
		require.NoError(t, err)

		// Check if all voted validators are returned
		valFoundMap := make(map[string]bool)
		for _, pk := range resp.BtcPks {
			if _, ok := votedValMap[pk.MarshalHex()]; !ok {
				t.Fatalf("rpc returned a val that was not created")
			}
			valFoundMap[pk.MarshalHex()] = true
		}
		if len(valFoundMap) != len(votedValMap) {
			t.Errorf("Some vals were missed. Got %d while %d were expected", len(valFoundMap), len(votedValMap))
		}
	})
}

func FuzzQueryEvidence(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// set random BTC SK PK
		sk, _, err := datagen.GenRandomBTCKeyPair(r)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey())
		require.NoError(t, err)

		var randomFirstSlashableEvidence *types.Evidence = nil
		numEvidences := datagen.RandomInt(r, 10) + 1
		height := uint64(5)

		// set a list of evidences, in which some of them are slashable while the others are not
		for i := uint64(0); i < numEvidences; i++ {
			evidence, err := datagen.GenRandomEvidence(r, sk, height)
			require.NoError(t, err)
			if datagen.RandomInt(r, 2) == 1 {
				evidence.CanonicalFinalitySig = nil // not slashable
			} else {
				if randomFirstSlashableEvidence == nil {
					randomFirstSlashableEvidence = evidence // first slashable
				}
			}
			keeper.SetEvidence(ctx, evidence)

			height += datagen.RandomInt(r, 5) + 1
		}

		// get first slashable evidence
		evidenceResp, err := keeper.Evidence(ctx, &types.QueryEvidenceRequest{ValBtcPkHex: bip340PK.MarshalHex()})
		if randomFirstSlashableEvidence == nil {
			require.Error(t, err)
			require.Nil(t, evidenceResp)
		} else {
			require.NoError(t, err)
			require.Equal(t, randomFirstSlashableEvidence, evidenceResp.Evidence)
			require.True(t, evidenceResp.Evidence.IsSlashable())
		}
	})
}
