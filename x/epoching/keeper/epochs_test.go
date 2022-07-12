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

		_, ctx, keeper, _, _ := setupTestKeeper()
		// ensure that the epoch info is correct at the genesis
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, epoch.EpochNumber, uint64(0))
		require.Equal(t, epoch.FirstBlockHeight, uint64(0))

		// set a random epoch interval
		epochInterval := rand.Uint64()%100 + 1
		keeper.SetParams(ctx, types.Params{
			EpochInterval: epochInterval,
		})
		// increment a random number of epochs
		numIncEpochs := rand.Uint64()%1000 + 1
		for i := uint64(0); i < numIncEpochs; i++ {
			keeper.IncEpoch(ctx)
		}

		// ensure that the epoch info is still correct
		newEpoch := keeper.GetEpoch(ctx)
		require.Equal(t, numIncEpochs, newEpoch.EpochNumber)
		require.Equal(t, epochInterval, newEpoch.CurrentEpochInterval)
		require.Equal(t, (numIncEpochs-1)*epochInterval+1, newEpoch.FirstBlockHeight)
	})
}
