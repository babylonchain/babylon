package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/babylonchain/babylon/x/checkpointing/types"
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
		ckptWithMeta, err := types.BytesToCkptWithMeta(k.cdc, value)
		if err != nil {
			return false, err
		}
		if ckptWithMeta.Status == req.Status {
			if accumulate {
				checkpointList = append(checkpointList, ckptWithMeta)
			}
			return true, nil
		}
		return false, nil
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

// RawCheckpoints returns checkpoints for given list of epoch numbers
func (k Keeper) RawCheckpoints(ctx context.Context, req *types.QueryRawCheckpointsRequest) (*types.QueryRawCheckpointsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.CheckpointsState(sdkCtx).checkpoints

	var checkpointList []*types.RawCheckpointWithMeta
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		ckptWithMeta, err := types.BytesToCkptWithMeta(k.cdc, value)
		if err != nil {
			return err
		}
		checkpointList = append(checkpointList, ckptWithMeta)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryRawCheckpointsResponse{RawCheckpoints: checkpointList, Pagination: pageRes}, nil
}

// EpochStatus returns the status of the checkpoint at a given epoch
func (k Keeper) EpochStatus(ctx context.Context, req *types.QueryEpochStatusRequest) (*types.QueryEpochStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ckptWithMeta, err := k.CheckpointsState(sdkCtx).GetRawCkptWithMeta(req.EpochNum)
	if err != nil {
		return nil, err
	}

	return &types.QueryEpochStatusResponse{Status: ckptWithMeta.Status}, nil
}

// RecentEpochStatusCount returns the count of epochs with each status of the checkpoint
func (k Keeper) RecentEpochStatusCount(ctx context.Context, req *types.QueryRecentEpochStatusCountRequest) (*types.QueryRecentEpochStatusCountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tipEpoch, err := k.GetLastCheckpointedEpoch(sdkCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get the last checkpointed epoch")
	}
	targetEpoch := tipEpoch - req.EpochCount + 1
	if targetEpoch < 0 { //nolint:staticcheck // uint64 doesn't go below zero
		targetEpoch = 0
	}
	// iterate epochs in the reverse order and count epoch numbers for each status
	epochStatusCount := make(map[string]uint64, 0)
	for e := tipEpoch; e >= targetEpoch; e-- {
		// reuse the EpochStatus query
		epochStatusReq := &types.QueryEpochStatusRequest{EpochNum: e}
		epochStatusRes, err := k.EpochStatus(ctx, epochStatusReq)
		if err != nil {
			return nil, err
		}
		// counts stop if a finalized epoch is reached since all the previous epochs are guaranteed to be finalized
		epochStatusCount[epochStatusRes.Status.String()]++
	}

	return &types.QueryRecentEpochStatusCountResponse{
		TipEpoch:    tipEpoch,
		EpochCount:  tipEpoch - targetEpoch + 1,
		StatusCount: epochStatusCount,
	}, nil
}

// LastCheckpointWithStatus returns the last checkpoint with the given status
// if the checkpoint with the given status does not exist, return the last
// checkpoint that is more mature than the given status
func (k Keeper) LastCheckpointWithStatus(ctx context.Context, req *types.QueryLastCheckpointWithStatusRequest) (*types.QueryLastCheckpointWithStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tipCheckpointedEpoch, err := k.GetLastCheckpointedEpoch(sdkCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get the last checkpointed epoch number: %w", err)
	}
	for e := int(tipCheckpointedEpoch); e >= 0; e-- {
		ckpt, err := k.GetRawCheckpoint(sdkCtx, uint64(e))
		if err != nil {
			return nil, fmt.Errorf("failed to get the raw checkpoint at epoch %v: %w", e, err)
		}
		if ckpt.Status == req.Status || ckpt.IsMoreMatureThanStatus(req.Status) {
			return &types.QueryLastCheckpointWithStatusResponse{RawCheckpoint: ckpt.Ckpt}, nil
		}
	}
	return nil, fmt.Errorf("cannot find checkpoint with status %v", req.Status)
}

// GetLastCheckpointedEpoch returns the last epoch number that associates with a checkpoint
func (k Keeper) GetLastCheckpointedEpoch(ctx sdk.Context) (uint64, error) {
	curEpoch := k.GetEpoch(ctx).EpochNumber
	if curEpoch <= 0 {
		return 0, fmt.Errorf("current epoch should be more than 0")
	}
	// minus 1 is because the current epoch is not ended
	tipEpoch := curEpoch - 1
	_, err := k.GetRawCheckpoint(ctx, tipEpoch)
	if err != nil {
		return 0, fmt.Errorf("cannot get raw checkpoint at epoch %v", tipEpoch)
	}
	return tipEpoch, nil
}
