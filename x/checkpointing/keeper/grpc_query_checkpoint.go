package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RawCheckpoints(c context.Context, req *types.QueryRawCheckpointsRequest) (*types.QueryRawCheckpointsResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) RecentRawCheckpoints(c context.Context, req *types.QueryRecentRawCheckpointsRequest) (*types.QueryRecentRawCheckpointsResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) RawCheckpoint(c context.Context, req *types.QueryRawCheckpointRequest) (*types.QueryRawCheckpointResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) LatestCheckpoint(c context.Context, req *types.QueryLatestCheckpointRequest) (*types.QueryLatestCheckpointResponse, error) {
	panic("TODO: implement this")
}
