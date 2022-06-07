package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RawCheckpoints(c context.Context, req *types.QueryRawCheckpointsRequest) (*types.QueryRawCheckpointsResponse, error) {
	panic("TODO: implement this")
}
