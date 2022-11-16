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
		numHeaders, numForkHeaders := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID)

		// check if the chain info is updated or not
		forks := zcKeeper.GetForks(ctx, czChain.ChainID, numHeaders-1)
		require.Equal(t, numForkHeaders, uint64(len(forks.Headers)))
	})
}
