package keeper_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

/*
	FuzzKeeperAddRawCheckpoint checks
	1. if the RawCheckpointWithMeta object is nil, an error is returned
	2. if the RawCheckpointWithMeta object does not exist when GetRawCheckpoint is called, an error is returned
	3. if a RawCheckpointWithMeta object with the same epoch number already exists, an error is returned
*/
func FuzzKeeperAddRawCheckpoint(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t)

		// test nil raw checkpoint
		err := ckptKeeper.AddRawCheckpoint(ctx, nil)
		require.Errorf(t, err, "add a nil raw checkpoint")

		// test random raw checkpoint
		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		ckpt, err := ckptKeeper.GetRawCheckpoint(ctx, mockCkptWithMeta.Ckpt.EpochNum)
		require.Nil(t, ckpt)
		require.Errorf(t, err, "raw checkpoint does not exist")
		err = ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		require.NoError(t, err)
		ckpt, err = ckptKeeper.GetRawCheckpoint(ctx, mockCkptWithMeta.Ckpt.EpochNum)
		require.NoError(t, err)
		t.Logf("mocked ckpt: %v\n got ckpt: %v\n", mockCkptWithMeta, ckpt)
		require.True(t, ckpt.Equal(mockCkptWithMeta))

		// test existing raw checkpoint by epoch number
		err = ckptKeeper.BuildRawCheckpoint(
			ctx,
			mockCkptWithMeta.Ckpt.EpochNum,
			datagen.GenRandomLastCommitHash(),
		)
		require.Errorf(t, err, "raw checkpoint with the same epoch already exists")
	})
}

/*
	FuzzKeeperCheckpointEpoch checks
	1. the rawCheckpointBytes is not valid, (0, err) should be returned
	2. the rawCheckpointBytes is valid, the correct epoch number should be returned
*/
func FuzzKeeperCheckpointEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, cdc := testkeeper.CheckpointingKeeper(t)

		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		ckptBytes := types.RawCkptToBytes(cdc, mockCkptWithMeta.Ckpt)
		epoch, err := ckptKeeper.CheckpointEpoch(ctx, ckptBytes)
		require.Equal(t, uint64(0), epoch)
		require.Errorf(t, err, "invalid checkpoint bytes")
		_ = ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		epoch, err = ckptKeeper.CheckpointEpoch(ctx, ckptBytes)
		require.NoError(t, err)
		require.Equal(t, mockCkptWithMeta.Ckpt.EpochNum, epoch)
	})
}

/*
	FuzzKeeperSetCheckpointStatus checks
	1. if the checkpoint is not valid, an error should be returned
	2. if the checkpoint is ACCUMULATING, an error should be returned
	2. the rawCheckpointBytes is valid, the correct epoch number should be returned
*/
func FuzzKeeperSetCheckpointSubmitted(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, cdc := testkeeper.CheckpointingKeeper(t)

		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		ckptBytes := types.RawCkptToBytes(cdc, mockCkptWithMeta.Ckpt)
		err := ckptKeeper.SetCheckpointSubmitted(ctx, ckptBytes)
		require.Errorf(t, err, "invalid checkpoint bytes")

		_ = ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		err = ckptKeeper.SetCheckpointSubmitted(ctx, ckptBytes)
		require.Errorf(t, err, "checkpoint status should be SEALED")
		mockCkptWithMeta.Status = types.Sealed
		err = ckptKeeper.UpdateCheckpoint(ctx, mockCkptWithMeta)
		require.NoError(t, err)
		err = ckptKeeper.SetCheckpointSubmitted(ctx, ckptBytes)
		require.NoError(t, err)
		err = ckptKeeper.SetCheckpointConfirmed(ctx, ckptBytes)
		require.NoError(t, err)
		err = ckptKeeper.SetCheckpointConfirmed(ctx, ckptBytes)
		require.Errorf(t, err, "checkpoint status should be SUBMITTED")
		err = ckptKeeper.SetCheckpointFinalized(ctx, ckptBytes)
		require.NoError(t, err)
		err = ckptKeeper.SetCheckpointFinalized(ctx, ckptBytes)
		require.Errorf(t, err, "checkpoint status should be CONFIRMED")
	})
}
