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
	var hashes []bbl.BTCHeaderHashBytes

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
	chHash, err := req.Hash.ToChainhash()
	if err != nil {
		return nil, err
	}
	contains, err := k.HeadersState(sdkCtx).HeaderExists(chHash)
	if err != nil {
		return nil, err
	}
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
	// If a starting key has not been set, then the first header is the tip
	prevHeader, err := k.HeadersState(sdkCtx).GetTip()
	if err != nil {
		return nil, err
	}
	// Otherwise, retrieve the header from the key
	if len(req.Pagination.Key) != 0 {
		headerHash := bbl.NewBTCHeaderHashBytesFromBytes(req.Pagination.Key)
		chHash, err := headerHash.ToChainhash()
		if err != nil {
			return nil, err
		}
		prevHeader, err = k.HeadersState(sdkCtx).GetHeaderByHash(chHash)
	}

	// If no tip exists or a key, then return an empty response
	if prevHeader == nil {
		return &types.QueryMainChainResponse{}, nil
	}

	var headers []*types.HeaderInfo
	headerInfo, err := types.NewHeaderInfo(prevHeader)
	if err != nil {
		return nil, err
	}
	headers = append(headers, headerInfo)
	store := prefix.NewStore(k.HeadersState(sdkCtx).headers, types.HeadersObjectPrefix)

	// Set this value to true to signal to FilteredPaginate to iterate the entries in reverse
	req.Pagination.Reverse = true
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		if accumulate {
			currentHeaderBytes, err := bbl.NewBTCHeaderBytesFromBytes(value)
			if err != nil {
				return false, err
			}
			btcdHeader, err := currentHeaderBytes.ToBlockHeader()
			if err != nil {
				return false, err
			}
			// If the previous block extends this block, then this block is part of the main chain
			if prevHeader.PrevBlock.String() == btcdHeader.BlockHash().String() {
				prevHeader = btcdHeader
				headerInfo, err := types.NewHeaderInfo(btcdHeader)
				if err != nil {
					return false, err
				}
				headers = append(headers, headerInfo)
			}
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	// Override the next key attribute to point to the parent of the last header
	// instead of the next element contained in the store
	prevBlockCh := prevHeader.PrevBlock
	pageRes.NextKey, err = bbl.NewBTCHeaderHashBytesFromChainhash(&prevBlockCh)
	if err != nil {
		return nil, err
	}

	return &types.QueryMainChainResponse{Headers: headers, Pagination: pageRes}, nil
}
