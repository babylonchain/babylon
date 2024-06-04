package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
)

func FuzzBlock(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		height := datagen.RandomInt(r, 100)
		appHash := datagen.GenRandomByteArray(r, 32)
		ib := &types.IndexedBlock{
			Height:  height,
			AppHash: appHash,
		}

		if datagen.RandomInt(r, 2) == 1 {
			ib.Finalized = true
		}

		keeper.SetBlock(ctx, ib)
		req := &types.QueryBlockRequest{
			Height: height,
		}
		resp, err := keeper.Block(ctx, req)
		require.NoError(t, err)
		require.Equal(t, height, resp.Block.Height)
		require.Equal(t, appHash, resp.Block.AppHash)
	})
}

func FuzzListBlocks(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// index a random list of finalised blocks
		startHeight := datagen.RandomInt(r, 100)
		numIndexedBlocks := datagen.RandomInt(r, 100) + 1
		finalizedIndexedBlocks := make(map[uint64]*types.IndexedBlock)
		nonFinalizedIndexedBlocks := make(map[uint64]*types.IndexedBlock)
		indexedBlocks := make(map[uint64]*types.IndexedBlock)
		for i := startHeight; i < startHeight+numIndexedBlocks; i++ {
			ib := &types.IndexedBlock{
				Height:  i,
				AppHash: datagen.GenRandomByteArray(r, 32),
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
				require.Equal(t, finalizedIndexedBlocks[actualIB.Height].AppHash, actualIB.AppHash)
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
				require.Equal(t, nonFinalizedIndexedBlocks[actualIB.Height].AppHash, actualIB.AppHash)
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
			require.Equal(t, indexedBlocks[actualIB.Height].AppHash, actualIB.AppHash)
		}
	})
}

func FuzzVotesAtHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// Add random number of voted finality providers to the store
		babylonHeight := datagen.RandomInt(r, 10) + 1
		numVotedFps := datagen.RandomInt(r, 10) + 1
		votedFpsMap := make(map[string]bool, numVotedFps)
		for i := uint64(0); i < numVotedFps; i++ {
			votedFpPK, err := datagen.GenRandomBIP340PubKey(r)
			require.NoError(t, err)
			votedSig, err := bbn.NewSchnorrEOTSSig(datagen.GenRandomByteArray(r, 32))
			require.NoError(t, err)
			keeper.SetSig(ctx, babylonHeight, votedFpPK, votedSig)

			votedFpsMap[votedFpPK.MarshalHex()] = true
		}

		resp, err := keeper.VotesAtHeight(ctx, &types.QueryVotesAtHeightRequest{
			Height: babylonHeight,
		})
		require.NoError(t, err)

		// Check if all voted finality providers are returned
		fpsFoundMap := make(map[string]bool)
		for _, pk := range resp.BtcPks {
			if _, ok := votedFpsMap[pk.MarshalHex()]; !ok {
				t.Fatalf("rpc returned a finality provider that was not created")
			}
			fpsFoundMap[pk.MarshalHex()] = true
		}
		if len(fpsFoundMap) != len(votedFpsMap) {
			t.Errorf("Some finality providers were missed. Got %d while %d were expected", len(fpsFoundMap), len(votedFpsMap))
		}
	})
}

func FuzzListPubRandCommit(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup keeper and context
		bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
		fKeeper, ctx := testkeeper.FinalityKeeper(t, bsKeeper, nil)
		ctx = sdk.UnwrapSDKContext(ctx)
		ms := keeper.NewMsgServerImpl(*fKeeper)

		// set random BTC SK PK
		sk, _, err := datagen.GenRandomBTCKeyPair(r)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey())
		require.NoError(t, err)

		// register finality provider
		fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, sk)
		require.NoError(t, err)
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(bip340PK.MustMarshal())).Return(fp, nil).AnyTimes()
		bsKeeper.EXPECT().HasFinalityProvider(gomock.Any(), gomock.Eq(bip340PK.MustMarshal())).Return(true).AnyTimes()

		numPrCommitList := datagen.RandomInt(r, 10) + 1
		prCommitList := []*types.PubRandCommit{}

		// set a list of random public randomness commitment
		startHeight := datagen.RandomInt(r, 10) + 1
		for i := uint64(0); i < numPrCommitList; i++ {
			numPubRand := datagen.RandomInt(r, 10) + 100
			randListInfo, err := datagen.GenRandomPubRandList(r, numPubRand)
			require.NoError(t, err)
			prCommit := &types.PubRandCommit{
				StartHeight: startHeight,
				NumPubRand:  numPubRand,
				Commitment:  randListInfo.Commitment,
			}
			msg := &types.MsgCommitPubRandList{
				Signer:      datagen.GenRandomAccount().Address,
				FpBtcPk:     bip340PK,
				StartHeight: startHeight,
				NumPubRand:  numPubRand,
				Commitment:  prCommit.Commitment,
			}
			hash, err := msg.HashToSign()
			require.NoError(t, err)
			schnorrSig, err := schnorr.Sign(sk, hash)
			require.NoError(t, err)
			msg.Sig = bbn.NewBIP340SignatureFromBTCSig(schnorrSig)
			_, err = ms.CommitPubRandList(ctx, msg)
			require.NoError(t, err)

			prCommitList = append(prCommitList, prCommit)

			startHeight += numPubRand
		}

		resp, err := fKeeper.ListPubRandCommit(ctx, &types.QueryListPubRandCommitRequest{
			FpBtcPkHex: bip340PK.MarshalHex(),
		})
		require.NoError(t, err)

		for _, prCommit := range prCommitList {
			prCommitResp, ok := resp.PubRandCommitMap[prCommit.StartHeight]
			require.True(t, ok)
			require.Equal(t, prCommitResp.NumPubRand, prCommit.NumPubRand)
			require.Equal(t, prCommitResp.Commitment, prCommit.Commitment)
		}
	})
}

func FuzzQueryEvidence(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil, nil)
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
		evidenceResp, err := keeper.Evidence(ctx, &types.QueryEvidenceRequest{FpBtcPkHex: bip340PK.MarshalHex()})
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

func FuzzListEvidences(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// Setup keeper and context
		keeper, ctx := testkeeper.FinalityKeeper(t, nil, nil)
		ctx = sdk.UnwrapSDKContext(ctx)

		// generate a random list of evidences since startHeight
		startHeight := datagen.RandomInt(r, 1000) + 100
		numEvidences := datagen.RandomInt(r, 100) + 10
		evidences := map[string]*types.Evidence{}
		for i := uint64(0); i < numEvidences; i++ {
			// random key pair
			sk, pk, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			btcPK := bbn.NewBIP340PubKeyFromBTCPK(pk)
			// random height
			height := datagen.RandomInt(r, 100) + startHeight + 1
			// generate evidence
			evidence, err := datagen.GenRandomEvidence(r, sk, height)
			require.NoError(t, err)
			// add evidence to map and finlaity keeper
			evidences[btcPK.MarshalHex()] = evidence
			keeper.SetEvidence(ctx, evidence)
		}

		// generate another list of evidences before startHeight
		// these evidences will not be included in the response if
		// the request specifies the above startHeight
		for i := uint64(0); i < numEvidences; i++ {
			// random key pair
			sk, _, err := datagen.GenRandomBTCKeyPair(r)
			require.NoError(t, err)
			// random height before startHeight
			height := datagen.RandomInt(r, int(startHeight))
			// generate evidence
			evidence, err := datagen.GenRandomEvidence(r, sk, height)
			require.NoError(t, err)
			// add evidence to finlaity keeper
			keeper.SetEvidence(ctx, evidence)
		}

		// perform a query to fetch all evidences and assert consistency
		limit := datagen.RandomInt(r, int(numEvidences)) + 1
		req := &types.QueryListEvidencesRequest{
			StartHeight: startHeight,
			Pagination: &query.PageRequest{
				CountTotal: true,
				Limit:      limit,
			},
		}
		resp, err := keeper.ListEvidences(ctx, req)
		require.NoError(t, err)
		require.LessOrEqual(t, len(resp.Evidences), int(limit))     // check if pagination takes effect
		require.EqualValues(t, resp.Pagination.Total, numEvidences) // ensure evidences before startHeight are not included
		for _, actualEvidence := range resp.Evidences {
			require.Equal(t, evidences[actualEvidence.FpBtcPk.MarshalHex()].CanonicalAppHash, actualEvidence.CanonicalAppHash)
			require.Equal(t, evidences[actualEvidence.FpBtcPk.MarshalHex()].ForkAppHash, actualEvidence.ForkAppHash)
		}
	})
}
