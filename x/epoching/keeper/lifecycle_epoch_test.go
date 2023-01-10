package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpochLifecycle(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelper(t)
		ctx, keeper := helper.Ctx, helper.EpochingKeeper
		hooks := keeper.Hooks()
		// ensure that the epoch info is correct at the genesis
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, epoch.EpochNumber, uint64(0))
		require.Equal(t, epoch.FirstBlockHeight, uint64(0))

		// set a random epoch interval
		epochInterval := rand.Uint64()%100 + 2 // the epoch interval should at at least 2
		keeper.SetParams(ctx, types.Params{
			EpochInterval: epochInterval,
		})

		// enter the last block of epoch 1
		for i := uint64(0); i < epochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		// ensure epoch 1 lifecycle contains correct started/ended state update
		curHeight := epochInterval
		lc := keeper.GetEpochLifecycle(ctx, 1)
		require.Len(t, lc.EpochLife, 2) // started and ended
		require.Equal(t, types.EpochState_STARTED, lc.EpochLife[0].State)
		require.Equal(t, uint64(1), lc.EpochLife[0].BlockHeight)
		require.Equal(t, types.EpochState_ENDED, lc.EpochLife[1].State)
		require.Equal(t, curHeight, lc.EpochLife[1].BlockHeight)

		// sealed
		numNewBlocks := datagen.RandomInt(10) + 1
		for i := uint64(0); i < numNewBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		curHeight += numNewBlocks
		hooks.AfterRawCheckpointSealed(ctx, 1)
		lc = keeper.GetEpochLifecycle(ctx, 1)
		require.Len(t, lc.EpochLife, 3) // started, ended, sealed
		require.Equal(t, types.EpochState_SEALED, lc.EpochLife[2].State)
		require.Equal(t, curHeight, lc.EpochLife[2].BlockHeight)

		// submitted
		numNewBlocks = datagen.RandomInt(10) + 1
		for i := uint64(0); i < numNewBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		curHeight += numNewBlocks
		hooks.AfterRawCheckpointSubmitted(ctx, 1)
		lc = keeper.GetEpochLifecycle(ctx, 1)
		require.Len(t, lc.EpochLife, 4) // started, ended, sealed, submitted
		require.Equal(t, types.EpochState_SUBMITTED, lc.EpochLife[3].State)
		require.Equal(t, curHeight, lc.EpochLife[3].BlockHeight)

		// confirmed
		numNewBlocks = datagen.RandomInt(10) + 1
		for i := uint64(0); i < numNewBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		curHeight += numNewBlocks
		hooks.AfterRawCheckpointConfirmed(ctx, 1)
		lc = keeper.GetEpochLifecycle(ctx, 1)
		require.Len(t, lc.EpochLife, 5) // started, ended, sealed, submitted, confirmed
		require.Equal(t, types.EpochState_CONFIRMED, lc.EpochLife[4].State)
		require.Equal(t, curHeight, lc.EpochLife[4].BlockHeight)

		// finalised
		numNewBlocks = datagen.RandomInt(10) + 1
		for i := uint64(0); i < numNewBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		curHeight += numNewBlocks
		hooks.AfterRawCheckpointFinalized(ctx, 1)
		lc = keeper.GetEpochLifecycle(ctx, 1)
		require.Len(t, lc.EpochLife, 6) // started, ended, sealed, submitted, confirmed, finalised
		require.Equal(t, types.EpochState_FINALIZED, lc.EpochLife[5].State)
		require.Equal(t, curHeight, lc.EpochLife[5].BlockHeight)
	})
}
