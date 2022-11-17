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

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		headers := SimulateHeadersViaHook(ctx, hooks, czChain.ChainID, numHeaders)

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
		chainInfo := zcKeeper.GetChainInfo(ctx, czChain.ChainID)
		require.NotNil(t, chainInfo.LatestHeader)
		require.Equal(t, czChain.ChainID, chainInfo.LatestHeader.ChainId)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, headers[numHeaders-1].Header.LastCommitHash, chainInfo.LatestHeader.Hash)
	})
}
