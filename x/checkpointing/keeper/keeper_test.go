package keeper_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func FuzzKeeperAddRawCheckpoint(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx := testkeeper.CheckpointingKeeper(t)
		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()

		err := ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		require.NoError(t, err)
		ckpt, err := ckptKeeper.GetRawCheckpoint(ctx, mockCkptWithMeta.Ckpt.EpochNum)
		require.True(t, ckpt.Equal(mockCkptWithMeta))
		err = ckptKeeper.BuildRawCheckpoint(
			ctx,
			mockCkptWithMeta.Ckpt.EpochNum,
			datagen.GenRandomLastCommitHash(),
		)
		require.Error(t, err)
	})
}
