package keeper

import (
	"context"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Hashes(ctx context.Context, req *types.QueryHashesRequest) (*types.QueryHashesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	var hashes bbl.BTCHeaderHashesBytes

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	store := prefix.NewStore(k.HeadersState(sdkCtx).hashToHeight, types.HashToHeightPrefix)
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, _ []byte, accumulate bool) (bool, error) {
		if accumulate {
			hashes = append(hashes, key)
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryHashesResponse{Hashes: hashes, Pagination: pageRes}, nil
}

func (k Keeper) Contains(ctx context.Context, req *types.QueryContainsRequest) (*types.QueryContainsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	chHash, err := req.Hash.MarshalChainhash()
	if err != nil {
		return nil, err
	}
	contains := k.HeadersState(sdkCtx).HeaderExists(chHash)
	return &types.QueryContainsResponse{Contains: contains}, nil
}
