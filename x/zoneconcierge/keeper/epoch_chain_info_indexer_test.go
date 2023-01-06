package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
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

		numReqs := datagen.RandomInt(5) + 1

		epochNumList := []uint64{datagen.RandomInt(10) + 1}
		nextHeightList := []uint64{0}
		numHeadersList := []uint64{}
		expectedHeadersMap := map[uint64][]*ibctmtypes.Header{}
		numForkHeadersList := []uint64{}

		// we test the scenario of ending an epoch for multiple times, in order to ensure that
		// consecutive epoch infos do not affect each other.
		for i := uint64(0); i < numReqs; i++ {
			// enter a random epoch
			for i := uint64(0); i < epochNumList[i]; i++ {
				epochingKeeper.IncEpoch(ctx)
			}

			// generate a random number of headers and fork headers
			numHeadersList = append(numHeadersList, datagen.RandomInt(100)+1)
			numForkHeadersList = append(numForkHeadersList, datagen.RandomInt(10)+1)
			// trigger hooks to append these headers and fork headers
			expectedHeaders, _ := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, nextHeightList[i], numHeadersList[i], numForkHeadersList[i])
			expectedHeadersMap[epochNumList[i]] = expectedHeaders
			// prepare nextHeight for the next request
			nextHeightList = append(nextHeightList, nextHeightList[i]+numHeadersList[i])

			// simulate the scenario that a random epoch has ended
			hooks.AfterEpochEnds(ctx, epochNumList[i])
			// prepare epochNum for the next request
			epochNumList = append(epochNumList, epochNumList[i]+datagen.RandomInt(10)+1)
		}

		// attest the correctness of epoch info for each tested epoch
		for i := uint64(0); i < numReqs; i++ {
			epochNum := epochNumList[i]
			// check if the headers are same as expected
			headers, err := zcKeeper.GetEpochHeaders(ctx, czChain.ChainID, epochNum)
			require.NoError(t, err)
			require.Equal(t, len(expectedHeadersMap[epochNum]), len(headers))
			for i := 0; i < len(expectedHeadersMap[epochNum]); i++ {
				require.Equal(t, expectedHeadersMap[epochNum][i].Header.LastCommitHash, headers[i].Hash)
			}
		}
	})
}
