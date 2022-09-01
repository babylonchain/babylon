package keeper_test

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/require"
)

func FuzzQueryEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})
		sdkCtx := sdk.WrapSDKContext(ctx)

		// test querying a raw checkpoint with epoch number
		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		err := ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		require.NoError(t, err)

		ckptRequest := types.NewQueryRawCheckpointRequest(mockCkptWithMeta.Ckpt.EpochNum)
		ckptResp, err := ckptKeeper.RawCheckpoint(sdkCtx, ckptRequest)
		require.NoError(t, err)
		require.True(t, ckptResp.RawCheckpoint.Equal(mockCkptWithMeta))

		// test querying the status of a given epoch number
		statusRequest := types.NewQueryEpochStatusRequest(mockCkptWithMeta.Ckpt.EpochNum)
		statusResp, err := ckptKeeper.EpochStatus(sdkCtx, statusRequest)
		require.NoError(t, err)
		require.Equal(t, mockCkptWithMeta.Status.String(), statusResp.Status)
	})
}

func FuzzQueryStatusCount(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})
		sdkCtx := sdk.WrapSDKContext(ctx)

		// test querying a raw checkpoint with epoch number
		checkpoints := datagen.GenRandomSequenceRawCheckpointsWithMeta()
		expectedCounts := make(map[string]uint64)
		for _, ckpt := range checkpoints {
			err := ckptKeeper.AddRawCheckpoint(
				ctx,
				ckpt,
			)
			require.NoError(t, err)
			expectedCounts[ckpt.Status.String()]++
		}

		countRequest := types.NewQueryEpochStatusCountRequest(uint64(len(checkpoints) - 1))
		resp, err := ckptKeeper.EpochStatusCount(sdkCtx, countRequest)
		require.NoError(t, err)
		require.Equal(t, expectedCounts, resp.StatusCount)
	})
}
