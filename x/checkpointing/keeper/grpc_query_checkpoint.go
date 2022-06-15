package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RawCheckpointList(c context.Context, req *types.QueryRawCheckpointListRequest) (*types.QueryRawCheckpointListResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) RecentRawCheckpointList(c context.Context, req *types.QueryRecentRawCheckpointListRequest) (*types.QueryRecentRawCheckpointListResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) RawCheckpoint(c context.Context, req *types.QueryRawCheckpointRequest) (*types.QueryRawCheckpointResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) LatestCheckpoint(c context.Context, req *types.QueryLatestCheckpointRequest) (*types.QueryLatestCheckpointResponse, error) {
	panic("TODO: implement this")
}
