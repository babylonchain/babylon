package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GetHashes(ctx context.Context, req *types.QueryGetHashesRequest) (*types.QueryGetHashesResponse, error) {
	var hashes [][]byte

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.HeadersState(sdkCtx).GetHeaders(func(hash []byte) bool {
		hashes = append(hashes, hash)
		// Return false which means that we want to continue receiving hashes
		return false
	})

	return &types.QueryGetHashesResponse{Hashes: hashes}, nil
}

func (k Keeper) ContainsHash(ctx context.Context, req *types.QueryContainsHashRequest) (*types.QueryContainsHashResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	containsHash := k.HeadersState(sdkCtx).Exists(req.Hash)
	return &types.QueryContainsHashResponse{ContainsHash: containsHash}, nil
}
