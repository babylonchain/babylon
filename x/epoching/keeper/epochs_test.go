package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/stretchr/testify/require"
)

func FuzzEpochs(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelper(t)
		ctx, keeper := helper.Ctx, helper.EpochingKeeper
		// ensure that the epoch info is correct at the genesis
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, epoch.EpochNumber, uint64(1))
		require.Equal(t, epoch.FirstBlockHeight, uint64(1))

		// set a random epoch interval
		epochInterval := keeper.GetParams(ctx).EpochInterval

		// increment a random number of new blocks
		numIncBlocks := r.Uint64()%1000 + 1
		var err error
		for i := uint64(0); i < numIncBlocks; i++ {
			// TODO: Figure out why when ctx height is 1, GenAndApplyEmptyBlock
			// will still give ctx height 1 once, then start to increment
			ctx, err = helper.GenAndApplyEmptyBlock(r)
			require.NoError(t, err)
		}

		// ensure that the epoch info is still correct
		expectedEpochNumber := (numIncBlocks + 1) / epochInterval
		if (numIncBlocks+1)%epochInterval > 0 {
			expectedEpochNumber += 1
		}
		actualNewEpoch := keeper.GetEpoch(ctx)
		require.Equal(t, expectedEpochNumber, actualNewEpoch.EpochNumber)
		require.Equal(t, epochInterval, actualNewEpoch.CurrentEpochInterval)
		require.Equal(t, (expectedEpochNumber-1)*epochInterval+1, actualNewEpoch.FirstBlockHeight)
	})
}
