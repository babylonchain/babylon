package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzCanonicalChainIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		headers := SimulateHeadersViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders)

		// check if the canonical chain index is correct or not
		for i := uint64(0); i < numHeaders; i++ {
			header, err := zcKeeper.GetHeader(ctx, czChain.ChainID, i)
			require.NoError(t, err)
			require.NotNil(t, header)
			require.Equal(t, czChain.ChainID, header.ChainId)
			require.Equal(t, i, header.Height)
			require.Equal(t, headers[i].Header.LastCommitHash, header.Hash)
		}

		// check if the chain info is updated or not
		chainInfo, err := zcKeeper.GetChainInfo(ctx, czChain.ChainID)
		require.NoError(t, err)
		require.NotNil(t, chainInfo.LatestHeader)
		require.Equal(t, czChain.ChainID, chainInfo.LatestHeader.ChainId)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, headers[numHeaders-1].Header.LastCommitHash, chainInfo.LatestHeader.Hash)
	})
}

func FuzzFindClosestHeader(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// no header at the moment, FindClosestHeader invocation should give error
		_, err := zcKeeper.FindClosestHeader(ctx, czChain.ChainID, 100)
		require.Error(t, err)

		// simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		headers := SimulateHeadersViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders)

		header, err := zcKeeper.FindClosestHeader(ctx, czChain.ChainID, numHeaders)
		require.NoError(t, err)
		require.Equal(t, headers[len(headers)-1].Header.LastCommitHash, header.Hash)

		// skip a non-zero number of headers in between, in order to create a gap of non-timestamped headers
		gap := datagen.RandomInt(10) + 1

		// simulate a random number of blocks
		// where the new batch of headers has a gap with the previous batch
		SimulateHeadersViaHook(ctx, hooks, czChain.ChainID, numHeaders+gap+1, numHeaders)

		// get a random height that is in this gap
		randomHeightInGap := datagen.RandomInt(int(gap+1)) + numHeaders
		// find the closest header with the given randomHeightInGap
		header, err = zcKeeper.FindClosestHeader(ctx, czChain.ChainID, randomHeightInGap)
		require.NoError(t, err)
		// the header should be the same as the last header in the last batch
		require.Equal(t, headers[len(headers)-1].Header.LastCommitHash, header.Hash)
	})
}
