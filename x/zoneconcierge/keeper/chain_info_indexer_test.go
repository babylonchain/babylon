package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzChainInfoIndexer_Canonical(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := SimulateHeadersViaHook(ctx, hooks, czChain.ChainID)

		// check if the chain info is updated or not
		chainInfo := zcKeeper.GetChainInfo(ctx, czChain.ChainID)
		require.NotNil(t, chainInfo.LatestHeader)
		require.Equal(t, czChain.ChainID, chainInfo.LatestHeader.ChainId)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
	})
}

func FuzzChainInfoIndexer_Fork(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		_, numForkHeaders := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID)

		// check if the chain info is updated or not
		chainInfo := zcKeeper.GetChainInfo(ctx, czChain.ChainID)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}
