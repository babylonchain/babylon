package keeper

import (
	"context"
	bbn "github.com/babylonchain/babylon/types"
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
	var hashes []bbn.BTCHeaderHashBytes

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Ensure that the pagination key corresponds to hash bytes
	if len(req.Pagination.Key) != 0 {
		_, err := bbn.NewBTCHeaderHashBytesFromBytes(req.Pagination.Key)
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

	var keyHeader *types.BTCHeaderInfo
	if len(req.Pagination.Key) != 0 {
		headerHash, err := bbn.NewBTCHeaderHashBytesFromBytes(req.Pagination.Key)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "key does not correspond to a header hash")
		}
		keyHeader, err = k.headersState(sdkCtx).GetHeaderByHash(&headerHash)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "header specified by key does not exist")
		}
	}

	var headers []*types.BTCHeaderInfo
	var nextKey []byte
	if req.Pagination.Reverse {
		var start, end uint64
		baseHeader := k.headersState(sdkCtx).GetBaseBTCHeader()
		// The base header is located at the end of the mainchain
		// which requires starting at the end
		mainchain := k.headersState(sdkCtx).GetMainChain()
		// Reverse the mainchain -- we want to retrieve results starting from the base header
		bbn.Reverse(mainchain)
		if keyHeader == nil {
			keyHeader = baseHeader
			start = 0
		} else {
			start = keyHeader.Height - baseHeader.Height
		}
		end = start + req.Pagination.Limit

		if end >= uint64(len(mainchain)) {
			end = uint64(len(mainchain))
		}

		// If the header's position on the mainchain is larger than the entire mainchain, then it is not part of the mainchain
		// Also, if the element at the header's position on the mainchain is not the provided one, then it is not part of the mainchain
		if start >= uint64(len(mainchain)) || !mainchain[start].Eq(keyHeader) {
			return nil, status.Error(codes.InvalidArgument, "header specified by key is not a part of the mainchain")
		}
		headers = mainchain[start:end]
		if end < uint64(len(mainchain)) {
			nextKey = mainchain[end].Hash.MustMarshal()
		}
	} else {
		tip := k.headersState(sdkCtx).GetTip()
		// If there is no starting key, then the starting header is the tip
		if keyHeader == nil {
			keyHeader = tip
		}
		// This is the depth in which the start header should in the mainchain
		startHeaderDepth := tip.Height - keyHeader.Height
		// The depth that we want to retrieve up to
		// -1 because the depth denotes how many headers have been built on top of it
		depth := startHeaderDepth + req.Pagination.Limit - 1
		// Retrieve the mainchain up to the depth
		mainchain := k.headersState(sdkCtx).GetMainChainUpTo(depth)
		// Check whether the key provided is part of the mainchain
		if uint64(len(mainchain)) <= startHeaderDepth || !mainchain[startHeaderDepth].Eq(keyHeader) {
			return nil, status.Error(codes.InvalidArgument, "header specified by key is not a part of the mainchain")
		}

		// The next key is the last elements parent hash
		nextKey = mainchain[len(mainchain)-1].Header.ParentHash().MustMarshal()
		headers = mainchain[startHeaderDepth:]
	}

	pageRes := &query.PageResponse{
		NextKey: nextKey,
	}
	// The headers that we should return start from the depth of the start header
	return &types.QueryMainChainResponse{Headers: headers, Pagination: pageRes}, nil
}

func (k Keeper) Tip(ctx context.Context, req *types.QueryTipRequest) (*types.QueryTipResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tip := k.headersState(sdkCtx).GetTip()

	return &types.QueryTipResponse{Header: tip}, nil
}
