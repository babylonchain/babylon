package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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
	var msgs []*types.QueuedMessage
	store := ctx.KVStore(k.storeKey)
	epochMsgsStore := prefix.NewStore(store, types.QueuedMsgKey)

	// handle pagination
	pageRes, err := query.Paginate(epochMsgsStore, req.Pagination, func(key, value []byte) error {
		// get key in the store
		storeKey := append(types.QueuedMsgKey, key...)
		// get value in bytes
		storeValue := store.Get(storeKey)
		// unmarshal to queuedMsg
		var queuedMsg types.QueuedMessage
		if err := k.cdc.Unmarshal(storeValue, &queuedMsg); err != nil {
			return err
		}
		// append to msgs
		msgs = append(msgs, &queuedMsg)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &types.QueryEpochMsgsResponse{
		Msgs:       msgs,
		Pagination: pageRes,
	}
	return resp, nil
}
