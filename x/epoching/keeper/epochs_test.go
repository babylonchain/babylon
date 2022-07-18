package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpochs(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		app, ctx, keeper, _, _, _ := SetupTestKeeper(t)
		// ensure that the epoch info is correct at the genesis
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, epoch.EpochNumber, uint64(0))
		require.Equal(t, epoch.FirstBlockHeight, uint64(0))

		// set a random epoch interval
		epochInterval := rand.Uint64()%100 + 1
		keeper.SetParams(ctx, types.Params{
			EpochInterval: epochInterval,
		})
		// increment a random number of new blocks
		numIncBlocks := rand.Uint64()%1000 + 1
		for i := uint64(0); i < numIncBlocks; i++ {
			ctx = genAndApplyEmptyBlock(app, ctx)
		}

		// ensure that the epoch info is still correct
		expectedEpochNumber := numIncBlocks / epochInterval
		if numIncBlocks%epochInterval > 0 {
			expectedEpochNumber += 1
		}
		actualNewEpoch := keeper.GetEpoch(ctx)
		require.Equal(t, expectedEpochNumber, actualNewEpoch.EpochNumber)
		require.Equal(t, epochInterval, actualNewEpoch.CurrentEpochInterval)
		require.Equal(t, (expectedEpochNumber-1)*epochInterval+1, actualNewEpoch.FirstBlockHeight)
	})
}
