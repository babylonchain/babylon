package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Hashes(ctx context.Context, req *types.QueryHashesRequest) (*types.QueryHashesResponse, error) {
	var hashes [][]byte

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.HeadersState(sdkCtx).GetBlockHashes(func(hash types.BlockHash) bool {
		hashes = append(hashes, hash)
		// Return false which means that we want to continue receiving hashes
		return false
	})

	return &types.QueryHashesResponse{Hashes: hashes}, nil
}

func (k Keeper) Contains(ctx context.Context, req *types.QueryContainsRequest) (*types.QueryContainsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	contains := k.HeadersState(sdkCtx).Exists(req.Hash)
	return &types.QueryContainsResponse{Contains: contains}, nil
}
