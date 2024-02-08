package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzCanonicalChainIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		babylonApp := app.Setup(t, false)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		ctx := babylonApp.NewContext(false)
		czChainID := "test-chainid"

		// simulate a random number of blocks
		numHeaders := datagen.RandomInt(r, 100) + 1
		headers := SimulateNewHeaders(ctx, r, &zcKeeper, czChainID, 0, numHeaders)

		// check if the canonical chain index is correct or not
		for i := uint64(0); i < numHeaders; i++ {
			header, err := zcKeeper.GetHeader(ctx, czChainID, i)
			require.NoError(t, err)
			require.NotNil(t, header)
			require.Equal(t, czChainID, header.ChainId)
			require.Equal(t, i, header.Height)
			require.Equal(t, headers[i].Header.AppHash, header.Hash)
		}

		// check if the chain info is updated or not
		chainInfo, err := zcKeeper.GetChainInfo(ctx, czChainID)
		require.NoError(t, err)
		require.NotNil(t, chainInfo.LatestHeader)
		require.Equal(t, czChainID, chainInfo.LatestHeader.ChainId)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, headers[numHeaders-1].Header.AppHash, chainInfo.LatestHeader.Hash)
	})
}

func FuzzFindClosestHeader(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		babylonApp := app.Setup(t, false)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		ctx := babylonApp.NewContext(false)
		czChainID := "test-chainid"

		// no header at the moment, FindClosestHeader invocation should give error
		_, err := zcKeeper.FindClosestHeader(ctx, czChainID, 100)
		require.Error(t, err)

		// simulate a random number of blocks
		numHeaders := datagen.RandomInt(r, 100) + 1
		headers := SimulateNewHeaders(ctx, r, &zcKeeper, czChainID, 0, numHeaders)

		header, err := zcKeeper.FindClosestHeader(ctx, czChainID, numHeaders)
		require.NoError(t, err)
		require.Equal(t, headers[len(headers)-1].Header.AppHash, header.Hash)

		// skip a non-zero number of headers in between, in order to create a gap of non-timestamped headers
		gap := datagen.RandomInt(r, 10) + 1

		// simulate a random number of blocks
		// where the new batch of headers has a gap with the previous batch
		SimulateNewHeaders(ctx, r, &zcKeeper, czChainID, numHeaders+gap+1, numHeaders)

		// get a random height that is in this gap
		randomHeightInGap := datagen.RandomInt(r, int(gap+1)) + numHeaders
		// find the closest header with the given randomHeightInGap
		header, err = zcKeeper.FindClosestHeader(ctx, czChainID, randomHeightInGap)
		require.NoError(t, err)
		// the header should be the same as the last header in the last batch
		require.Equal(t, headers[len(headers)-1].Header.AppHash, header.Hash)
	})
}
