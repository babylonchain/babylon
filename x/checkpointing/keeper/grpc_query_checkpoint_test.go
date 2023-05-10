package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/babylonchain/babylon/testutil/mocks"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
)

func FuzzQueryEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})
		sdkCtx := sdk.WrapSDKContext(ctx)

		// test querying a raw checkpoint with epoch number
		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta(r)
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

func FuzzQueryRawCheckpoints(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})
		sdkCtx := sdk.WrapSDKContext(ctx)

		// add a random number of checkpoints
		checkpoints := datagen.GenRandomSequenceRawCheckpointsWithMeta(r)
		for _, ckpt := range checkpoints {
			err := ckptKeeper.AddRawCheckpoint(
				ctx,
				ckpt,
			)
			require.NoError(t, err)
		}

		// test querying raw checkpoints with epoch range in pagination params
		startEpoch := checkpoints[0].Ckpt.EpochNum
		endEpoch := checkpoints[len(checkpoints)-1].Ckpt.EpochNum
		pageLimit := endEpoch - startEpoch + 1

		pagination := &query.PageRequest{Key: types.CkptsObjectKey(startEpoch), Limit: pageLimit}
		ckptResp, err := ckptKeeper.RawCheckpoints(sdkCtx, &types.QueryRawCheckpointsRequest{Pagination: pagination})
		require.NoError(t, err)
		require.Equal(t, int(pageLimit), len(ckptResp.RawCheckpoints))
		require.Nil(t, ckptResp.Pagination.NextKey)
		for i, ckpt := range ckptResp.RawCheckpoints {
			require.Equal(t, checkpoints[i], ckpt)
		}
	})
}

func FuzzQueryStatusCount(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// test querying recent epoch counts with each status in recent epochs
		checkpoints := datagen.GenRandomSequenceRawCheckpointsWithMeta(r)
		tipEpoch := checkpoints[len(checkpoints)-1].Ckpt.EpochNum
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ek.EXPECT().GetEpoch(gomock.Any()).Return(&epochingtypes.Epoch{EpochNumber: tipEpoch + 1})
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})
		sdkCtx := sdk.WrapSDKContext(ctx)
		expectedCounts := make(map[string]uint64)
		epochCount := uint64(r.Int63n(int64(tipEpoch)))
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

func FuzzQueryLastCheckpointWithStatus(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// test querying recent epoch counts with each status in recent epochs
		tipEpoch := datagen.RandomInt(r, 100) + 10
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ek.EXPECT().GetEpoch(gomock.Any()).Return(&epochingtypes.Epoch{EpochNumber: tipEpoch}).AnyTimes()
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})
		checkpoints := datagen.GenSequenceRawCheckpointsWithMeta(r, tipEpoch)
		finalizedEpoch := datagen.RandomInt(r, int(tipEpoch))
		for e := uint64(0); e < tipEpoch; e++ {
			if e <= finalizedEpoch {
				checkpoints[int(e)].Status = types.Finalized
			} else {
				checkpoints[int(e)].Status = types.Sealed
			}
			err := ckptKeeper.AddRawCheckpoint(ctx, checkpoints[int(e)])
			require.NoError(t, err)
		}
		// request the last finalized checkpoint
		req := types.NewQueryLastCheckpointWithStatus(types.Finalized)
		expectedResp := &types.QueryLastCheckpointWithStatusResponse{
			RawCheckpoint: checkpoints[int(finalizedEpoch)].Ckpt,
		}
		resp, err := ckptKeeper.LastCheckpointWithStatus(ctx, req)
		require.NoError(t, err)
		require.Equal(t, expectedResp, resp)

		// request the last confirmed checkpoint
		req = types.NewQueryLastCheckpointWithStatus(types.Confirmed)
		expectedResp = &types.QueryLastCheckpointWithStatusResponse{
			RawCheckpoint: checkpoints[int(finalizedEpoch)].Ckpt,
		}
		resp, err = ckptKeeper.LastCheckpointWithStatus(ctx, req)
		require.NoError(t, err)
		require.Equal(t, expectedResp, resp)
	})
}

// func TestQueryRawCheckpointList(t *testing.T) {
func FuzzQueryRawCheckpointList(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		tipEpoch := datagen.RandomInt(r, 10) + 10
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ek.EXPECT().GetEpoch(gomock.Any()).Return(&epochingtypes.Epoch{EpochNumber: tipEpoch}).AnyTimes()
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})
		checkpoints := datagen.GenSequenceRawCheckpointsWithMeta(r, tipEpoch)
		finalizedEpoch := datagen.RandomInt(r, int(tipEpoch))

		// add Sealed and Finalized checkpoints
		for e := uint64(0); e <= tipEpoch; e++ {
			if e <= finalizedEpoch {
				checkpoints[int(e)].Status = types.Finalized
			} else {
				checkpoints[int(e)].Status = types.Sealed
			}
			err := ckptKeeper.AddRawCheckpoint(ctx, checkpoints[int(e)])
			require.NoError(t, err)
		}

		finalizedCheckpoints := checkpoints[:finalizedEpoch+1]
		testRawCheckpointListWithType(t, r, ckptKeeper, ctx, finalizedCheckpoints, 0, types.Finalized)
		sealedCheckpoints := checkpoints[finalizedEpoch+1:]
		testRawCheckpointListWithType(t, r, ckptKeeper, ctx, sealedCheckpoints, finalizedEpoch+1, types.Sealed)
	})
}

func testRawCheckpointListWithType(
	t *testing.T,
	r *rand.Rand,
	ckptKeeper *keeper.Keeper,
	ctx context.Context,
	checkpointList []*types.RawCheckpointWithMeta,
	baseEpoch uint64,
	status types.CheckpointStatus,
) {
	limit := datagen.RandomInt(r, len(checkpointList)+1) + 1
	pagination := &query.PageRequest{Limit: limit, CountTotal: true}
	req := types.NewQueryRawCheckpointListRequest(pagination, status)

	resp, err := ckptKeeper.RawCheckpointList(ctx, req)
	require.NoError(t, err)
	require.Equal(t, uint64(len(checkpointList)), resp.Pagination.Total)
	for ckptsRetrieved := uint64(0); ckptsRetrieved < uint64(len(checkpointList)); ckptsRetrieved += limit {
		resp, err := ckptKeeper.RawCheckpointList(ctx, req)
		require.NoError(t, err)
		for i, ckpt := range resp.RawCheckpoints {
			require.Equal(t, baseEpoch+ckptsRetrieved+uint64(i), ckpt.Ckpt.EpochNum)
			require.Equal(t, status, ckpt.Status)
		}
		pagination = &query.PageRequest{Key: resp.Pagination.NextKey, Limit: limit}
		req = types.NewQueryRawCheckpointListRequest(pagination, status)
	}
}
