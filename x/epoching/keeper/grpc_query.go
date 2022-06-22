package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
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

// CurrentEpoch handles the QueryCurrentEpochRequest query
func (k Keeper) CurrentEpoch(c context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return nil, err
	}
	epochBoundary, err := k.GetEpochBoundary(ctx)
	if err != nil {
		return nil, err
	}
	resp := &types.QueryCurrentEpochResponse{
		CurrentEpoch:  epochNumber.BigInt().Uint64(),
		EpochBoundary: epochBoundary.BigInt().Uint64(),
	}
	return resp, nil
}

// EpochMsgs handles the QueryEpochMsgsRequest query
func (k Keeper) EpochMsgs(c context.Context, req *types.QueryEpochMsgsRequest) (*types.QueryEpochMsgsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	msgs, err := k.GetEpochMsgs(ctx)
	if err != nil {
		return nil, err
	}
	resp := &types.QueryEpochMsgsResponse{
		Msgs: msgs,
		Pagination: &qtypes.PageResponse{
			Total: uint64(len(msgs)),
		},
	}
	return resp, nil
}
