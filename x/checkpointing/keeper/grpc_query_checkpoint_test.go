package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/mocks"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

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
		require.Equal(t, mockCkptWithMeta.Status, statusResp.Status)
	})
}

func FuzzQueryStatusCount(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		// test querying recent epoch counts with each status in recent epochs
		checkpoints := datagen.GenRandomSequenceRawCheckpointsWithMeta()
		tipEpoch := checkpoints[len(checkpoints)-1].Ckpt.EpochNum
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ek.EXPECT().GetEpoch(gomock.Any()).Return(&epochingtypes.Epoch{EpochNumber: tipEpoch + 1})
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})
		sdkCtx := sdk.WrapSDKContext(ctx)
		expectedCounts := make(map[string]uint64)
		epochCount := uint64(rand.Int63n(int64(tipEpoch)))
		for e, ckpt := range checkpoints {
			err := ckptKeeper.AddRawCheckpoint(
				ctx,
				ckpt,
			)
			require.NoError(t, err)
			if uint64(e) >= tipEpoch-epochCount+1 {
				expectedCounts[ckpt.Status.String()]++
			}
		}
		expectedResp := &types.QueryRecentEpochStatusCountResponse{
			TipEpoch:    tipEpoch,
			EpochCount:  epochCount,
			StatusCount: expectedCounts,
		}

		countRequest := types.NewQueryRecentEpochStatusCountRequest(epochCount)
		resp, err := ckptKeeper.RecentEpochStatusCount(sdkCtx, countRequest)
		require.NoError(t, err)
		require.Equal(t, expectedResp, resp)
	})
}
