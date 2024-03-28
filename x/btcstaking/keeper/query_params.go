package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) ParamsByVersion(goCtx context.Context, req *types.QueryParamsByVersionRequest) (*types.QueryParamsByVersionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	pv := k.GetParamsByVersion(ctx, req.Version)

	if pv == nil {
		return nil, types.ErrParamsNotFound.Wrapf("version %d does not exists", req.Version)
	}

	return &types.QueryParamsByVersionResponse{Params: *pv}, nil
}
