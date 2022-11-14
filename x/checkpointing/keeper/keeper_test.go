package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/mocks"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
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
		_, err = ckptKeeper.BuildRawCheckpoint(
			ctx,
			mockCkptWithMeta.Ckpt.EpochNum,
			datagen.GenRandomLastCommitHash(),
		)
		require.Errorf(t, err, "raw checkpoint with the same epoch already exists")
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

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})

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
