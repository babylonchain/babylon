package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// RawCheckpointList returns a list of checkpoint by status in the ascending order of epoch
func (k Keeper) RawCheckpointList(ctx context.Context, req *types.QueryRawCheckpointListRequest) (*types.QueryRawCheckpointListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	var checkpointList []*types.RawCheckpointWithMeta

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	store := k.CheckpointsState(sdkCtx).checkpoints
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		if accumulate {
			ckptWithMeta, err := types.BytesToCkptWithMeta(k.cdc, value)
			if err != nil {
				return false, err
			}
			if ckptWithMeta.Status == req.Status {
				checkpointList = append(checkpointList, ckptWithMeta)
			}
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryRawCheckpointListResponse{RawCheckpoints: checkpointList, Pagination: pageRes}, nil
}

// RawCheckpoint returns a checkpoint by epoch number
func (k Keeper) RawCheckpoint(ctx context.Context, req *types.QueryRawCheckpointRequest) (*types.QueryRawCheckpointResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ckptWithMeta, err := k.CheckpointsState(sdkCtx).GetRawCkptWithMeta(req.EpochNum)
	if err != nil {
		return nil, err
	}

	return &types.QueryRawCheckpointResponse{RawCheckpoint: ckptWithMeta}, nil
}

func (k Keeper) EpochStatus(ctx context.Context, req *types.QueryEpochStatusRequest) (*types.QueryEpochStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ckptWithMeta, err := k.CheckpointsState(sdkCtx).GetRawCkptWithMeta(req.EpochNum)
	if err != nil {
		return nil, err
	}

	return &types.QueryEpochStatusResponse{Status: ckptWithMeta.Status.String()}, nil
}

func (k Keeper) EpochStatusCount(ctx context.Context, req *types.QueryEpochStatusCountRequest) (*types.QueryEpochStatusCountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// iterate epochs from 0 to the current epoch and count the status
	epochStatusCount := make(map[string]uint64, 0)
	for i := uint64(0); i <= req.EpochNum; i++ {
		// reuse the EpochStatus query
		epochStatusReq := &types.QueryEpochStatusRequest{EpochNum: i}
		epochStatusRes, err := k.EpochStatus(ctx, epochStatusReq)
		if err != nil {
			return nil, err
		}
		epochStatusCount[epochStatusRes.Status]++
	}

	return &types.QueryEpochStatusCountResponse{StatusCount: epochStatusCount}, nil
}

func (k Keeper) RecentRawCheckpointList(c context.Context, req *types.QueryRecentRawCheckpointListRequest) (*types.QueryRecentRawCheckpointListResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) LatestCheckpoint(c context.Context, req *types.QueryLatestCheckpointRequest) (*types.QueryLatestCheckpointResponse, error) {
	panic("TODO: implement this")
}
