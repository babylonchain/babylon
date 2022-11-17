package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzForkIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		_, forkHeaders := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, numHeaders, numForkHeaders)

		// check if the fork is updated or not
		forks := zcKeeper.GetForks(ctx, czChain.ChainID, numHeaders-1)
		require.Equal(t, numForkHeaders, uint64(len(forks.Headers)))
		for i := range forks.Headers {
			require.Equal(t, czChain.ChainID, forks.Headers[i].ChainId)
			require.Equal(t, numHeaders-1, forks.Headers[i].Height)
			require.Equal(t, forkHeaders[i].Header.LastCommitHash, forks.Headers[i].Hash)
		}

		// check if the chain info is updated or not
		chainInfo := zcKeeper.GetChainInfo(ctx, czChain.ChainID)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
		for i := range forks.Headers {
			require.Equal(t, czChain.ChainID, chainInfo.LatestForks.Headers[i].ChainId)
			require.Equal(t, numHeaders-1, chainInfo.LatestForks.Headers[i].Height)
			require.Equal(t, forkHeaders[i].Header.LastCommitHash, chainInfo.LatestForks.Headers[i].Hash)
		}
	})
}
