package keeper_test

import (
	"math/rand"
	"testing"

	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
)

func FuzzEpochChainInfoIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		babylonApp := app.Setup(t, false)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		ctx := babylonApp.NewContext(false)
		czChainID := "test-chainid"

		hooks := zcKeeper.Hooks()

		// enter a random epoch
		epochNum := datagen.RandomInt(r, 10)
		for j := uint64(0); j < epochNum; j++ {
			babylonApp.EpochingKeeper.IncEpoch(ctx)
		}

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(r, 100) + 1
		numForkHeaders := datagen.RandomInt(r, 10) + 1
		SimulateNewHeadersAndForks(ctx, r, &zcKeeper, czChainID, 0, numHeaders, numForkHeaders)

		// end this epoch
		hooks.AfterEpochEnds(ctx, epochNum)

		// check if the chain info of this epoch is recorded or not
		chainInfoWithProof, err := zcKeeper.GetEpochChainInfo(ctx, czChainID, epochNum)
		chainInfo := chainInfoWithProof.ChainInfo
		require.NoError(t, err)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numHeaders, chainInfo.TimestampedHeadersCount)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}

func FuzzGetEpochHeaders(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		babylonApp := app.Setup(t, false)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		ctx := babylonApp.NewContext(false)
		czChainID := "test-chainid"

		hooks := zcKeeper.Hooks()

		numReqs := datagen.RandomInt(r, 5) + 1
		epochNumList := []uint64{datagen.RandomInt(r, 10) + 1}
		nextHeightList := []uint64{0}
		numHeadersList := []uint64{}
		expectedHeadersMap := map[uint64][]*ibctmtypes.Header{}
		numForkHeadersList := []uint64{}

		// we test the scenario of ending an epoch for multiple times, in order to ensure that
		// consecutive epoch infos do not affect each other.
		for i := uint64(0); i < numReqs; i++ {
			epochNum := epochNumList[i]
			// enter a random epoch
			if i == 0 {
				for j := uint64(1); j < epochNum; j++ { // starting from epoch 1
					babylonApp.EpochingKeeper.IncEpoch(ctx)
				}
			} else {
				for j := uint64(0); j < epochNum-epochNumList[i-1]; j++ {
					babylonApp.EpochingKeeper.IncEpoch(ctx)
				}
			}

			// generate a random number of headers and fork headers
			numHeadersList = append(numHeadersList, datagen.RandomInt(r, 100)+1)
			numForkHeadersList = append(numForkHeadersList, datagen.RandomInt(r, 10)+1)
			// trigger hooks to append these headers and fork headers
			expectedHeaders, _ := SimulateNewHeadersAndForks(ctx, r, &zcKeeper, czChainID, nextHeightList[i], numHeadersList[i], numForkHeadersList[i])
			expectedHeadersMap[epochNum] = expectedHeaders
			// prepare nextHeight for the next request
			nextHeightList = append(nextHeightList, nextHeightList[i]+numHeadersList[i])

			// simulate the scenario that a random epoch has sealed
			err := hooks.AfterRawCheckpointSealed(ctx, epochNum)
			require.NoError(t, err)
			// prepare epochNum for the next request
			epochNumList = append(epochNumList, epochNum+datagen.RandomInt(r, 10)+1)
		}

		// attest the correctness of epoch info for each tested epoch
		for i := uint64(0); i < numReqs; i++ {
			epochNum := epochNumList[i]
			// check if the headers are same as expected
			headers, err := zcKeeper.GetEpochHeaders(ctx, czChainID, epochNum)
			require.NoError(t, err)
			require.Equal(t, len(expectedHeadersMap[epochNum]), len(headers))
			for j := 0; j < len(expectedHeadersMap[epochNum]); j++ {
				require.Equal(t, expectedHeadersMap[epochNum][j].Header.AppHash, headers[j].Hash)
			}
		}
	})
}
