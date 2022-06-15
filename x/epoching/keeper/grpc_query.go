package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) CurrentEpoch(c context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	panic("TODO: unimplemented")
}

func (k Keeper) EpochMsgs(c context.Context, req *types.QueryEpochMsgsRequest) (*types.QueryEpochMsgsResponse, error) {
	panic("TODO: unimplemented")
}

func (k Keeper) ValidatorBLSPK(c context.Context, req *types.QueryValidatorBLSPKRequest) (*types.QueryValidatorBLSPKResponse, error) {
	panic("TODO: unimplemented")
}
