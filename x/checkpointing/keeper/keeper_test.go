package keeper_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/client"
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
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})

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
		ckptKeeper, ctx, cdc := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})

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
	if the checkpoint status is not correct, the status will not be changed
*/
func FuzzKeeperSetCheckpointStatus(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})

		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		mockCkptWithMeta.Status = types.Accumulating
		epoch := mockCkptWithMeta.Ckpt.EpochNum

		_ = ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		ckptKeeper.SetCheckpointSubmitted(ctx, epoch)
		status, err := ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Accumulating, status)
		mockCkptWithMeta.Status = types.Sealed
		err = ckptKeeper.UpdateCheckpoint(ctx, mockCkptWithMeta)
		require.NoError(t, err)
		ckptKeeper.SetCheckpointSubmitted(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Submitted, status)
		ckptKeeper.SetCheckpointConfirmed(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Confirmed, status)
		ckptKeeper.SetCheckpointConfirmed(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Confirmed, status)
		ckptKeeper.SetCheckpointFinalized(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Finalized, status)
	})
}
