package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RawCheckpoints(c context.Context, req *types.QueryRawCheckpointsRequest) (*types.QueryRawCheckpointsResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) RawCheckpoint(c context.Context, req *types.QueryRawCheckpointRequest) (*types.QueryRawCheckpointResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) LatestCheckpoint(c context.Context, req *types.QueryLatestCheckpointRequest) (*types.QueryLatestCheckpointResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) UncheckpointedCheckpoints(c context.Context, req *types.QueryUncheckpointedCheckpointsRequest) (*types.QueryUncheckpointedCheckpointsResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) UnderconfirmedCheckpoints(c context.Context, req *types.QueryUnderconfirmedCheckpointsRequest) (*types.QueryUnderconfirmedCheckpointsResponse, error) {
	panic("TODO: implement this")
}

func (k Keeper) ConfirmedCheckpoints(c context.Context, req *types.QueryConfirmedCheckpointsRequest) (*types.QueryConfirmedCheckpointsResponse, error) {
	panic("TODO: implement this")
}
