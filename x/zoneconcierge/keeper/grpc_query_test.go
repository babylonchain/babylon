package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/stretchr/testify/require"
)

func FuzzFinalizedChainInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100)
		for i := uint64(0); i < numHeaders; i++ {
			header := datagen.GenRandomIBCTMHeader(czChain.ChainID, i)
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		}

		// generate a number of fork headers
		numForkHeaders := int(datagen.RandomInt(10))
		for i := 0; i < numForkHeaders; i++ {
			header := datagen.GenRandomIBCTMHeader(czChain.ChainID, numHeaders-1)
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, true)
		}

		// simulate the scenario that a random epoch has ended and finalised
		epochNum := datagen.RandomInt(10)
		hooks.AfterEpochEnds(ctx, epochNum)
		hooks.AfterRawCheckpointFinalized(ctx, epochNum)

		// check if the chain info of this epoch is recorded or not
		resp, err := zcKeeper.FinalizedChainInfo(ctx, &zctypes.QueryFinalizedChainInfoRequest{ChainId: czChain.ChainID})
		require.NoError(t, err)
		chainInfo := resp.FinalizedChainInfo
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, len(chainInfo.LatestForks.Headers))
	})
}
