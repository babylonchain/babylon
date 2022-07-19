package keeper

import (
	"context"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
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
	var hashes []bbl.BTCHeaderHashBytes

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Ensure that the pagination key corresponds to hash bytes
	if len(req.Pagination.Key) != 0 {
		_, err := bbl.NewBTCHeaderHashBytesFromBytes(req.Pagination.Key)
		if err != nil {
			return nil, err
		}
	}

	store := k.headersState(sdkCtx).hashToHeight
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
	contains := k.headersState(sdkCtx).HeaderExists(req.Hash)
	return &types.QueryContainsResponse{Contains: contains}, nil
}

func (k Keeper) MainChain(ctx context.Context, req *types.QueryMainChainRequest) (*types.QueryMainChainResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req.Pagination == nil {
		req.Pagination = &query.PageRequest{}
	}

	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = query.DefaultLimit
	}

	tip := k.headersState(sdkCtx).GetTip()
	var startHeader *types.BTCHeaderInfo
	if len(req.Pagination.Key) != 0 {
		headerHash, err := bbl.NewBTCHeaderHashBytesFromBytes(req.Pagination.Key)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "key does not correspond to a header hash")
		}
		startHeader, err = k.headersState(sdkCtx).GetHeaderByHash(&headerHash)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "header specified by key does not exist")
		}
	} else {
		startHeader = tip
	}

	// This is the depth in which the start header should in the mainchain
	startHeaderDepth := tip.Height - startHeader.Height
	// The depth that we want to retrieve up to
	// -1 because the depth denotes how many headers have been built on top of it
	depth := startHeaderDepth + req.Pagination.Limit - 1
	// Retrieve the mainchain up to the depth
	mainchain := k.headersState(sdkCtx).GetMainChainUpTo(depth)
	// Check whether the key provided is part of the mainchain
	if uint64(len(mainchain)) <= startHeaderDepth || !mainchain[startHeaderDepth].Eq(startHeader) {
		return nil, status.Error(codes.InvalidArgument, "header specified by key is not a part of the mainchain")
	}

	nextKey := mainchain[len(mainchain)-1].Header.ParentHash().MustMarshal()
	headers := mainchain[startHeaderDepth:]
	pageRes := &query.PageResponse{
		NextKey: nextKey,
	}
	return &types.QueryMainChainResponse{Headers: headers, Pagination: pageRes}, nil
}
