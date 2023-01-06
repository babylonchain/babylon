package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzEpochChainInfoIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders, numForkHeaders)

		// simulate the scenario that a random epoch has ended
		epochNum := datagen.RandomInt(10)
		hooks.AfterEpochEnds(ctx, epochNum)

		// check if the chain info of this epoch is recorded or not
		chainInfo, err := zcKeeper.GetEpochChainInfo(ctx, czChain.ChainID, epochNum)
		require.NoError(t, err)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}

func FuzzGetEpochHeaders(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		epochingKeeper := babylonApp.EpochingKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// enter a random epoch
		epochNum := datagen.RandomInt(10) + 1
		for i := uint64(0); i < epochNum; i++ {
			epochingKeeper.IncEpoch(ctx)
		}

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		expectedHeaders, _ := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders, numForkHeaders)

		// end this epoch so that the chain info is recorded
		hooks.AfterEpochEnds(ctx, epochNum)

		// check if the headers are same as expected
		headers, err := zcKeeper.GetEpochHeaders(ctx, czChain.ChainID, epochNum)
		require.NoError(t, err)
		require.Equal(t, len(expectedHeaders), len(headers))
		for i := 0; i < len(expectedHeaders); i++ {
			require.Equal(t, expectedHeaders[i].Header.LastCommitHash, headers[i].Hash)
		}
	})
}
