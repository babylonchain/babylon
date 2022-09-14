package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	epoch := k.GetEpoch(ctx)
	resp := &types.QueryCurrentEpochResponse{
		CurrentEpoch:  epoch.EpochNumber,
		EpochBoundary: epoch.GetLastBlockHeight(),
	}
	return resp, nil
}

// EpochMsgs handles the QueryEpochMsgsRequest query
func (k Keeper) EpochMsgs(c context.Context, req *types.QueryEpochMsgsRequest) (*types.QueryEpochMsgsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	epoch := k.GetEpoch(ctx)
	if epoch.EpochNumber < req.EpochNum {
		return nil, types.ErrUnknownEpochNumber
	}

	var msgs []*types.QueuedMessage
	epochMsgsStore := k.msgQueueStore(ctx, req.EpochNum)

	// handle pagination
	// TODO (non-urgent): the epoch might end between pagination requests, leading inconsistent results by the time someone gets to the end. Possible fixes:
	// - We could add the epoch number to the query, and return nothing if the current epoch number is different. But it's a bit of pain to have to set it and not know why there's no result.
	// - We could not reset the key to 0 when the queue is cleared, and just keep incrementing the ID forever. That way when the next query comes, it might skip some records that have been deleted, then resume from the next available record which has a higher key than the value in the pagination data structure.
	// - We can do nothing, in which case some records that have been inserted after the delete might be skipped because their keys are lower than the pagionation state.
	pageRes, err := query.Paginate(epochMsgsStore, req.Pagination, func(key, value []byte) error {
		// unmarshal to queuedMsg
		var queuedMsg types.QueuedMessage
		if err := k.cdc.Unmarshal(value, &queuedMsg); err != nil {
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

// LatestEpochMsgs handles the QueryLatestEpochMsgsRequest query
// TODO: test this API
func (k Keeper) LatestEpochMsgs(c context.Context, req *types.QueryLatestEpochMsgsRequest) (*types.QueryLatestEpochMsgsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if req.EpochCount == 0 {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "epoch_count should be specified and be larger than zero",
		)
	}

	// the API will return epoch msgs between [min(1, end_epoch-epoch_count+1), end_epoch].
	// NOTE:
	// - epoch 0 does not have any queued msg
	// - if not specified, endEpoch will be the current epoch
	endEpoch := req.EndEpoch
	if endEpoch == 0 {
		endEpoch = k.GetEpoch(ctx).EpochNumber
	}
	beginEpoch := endEpoch - req.EpochCount + 1
	if beginEpoch <= 1 {
		beginEpoch = 1
	}

	latestEpochMsgs := []*types.QueuedMessageList{}

	// iterate over queueLenStore since we only need to iterate over the epoch number
	queueLenStore := k.msgQueueLengthStore(ctx)
	pageRes, err := query.FilteredPaginate(queueLenStore, req.Pagination, func(key []byte, _ []byte, accumulate bool) (bool, error) {
		// unmarshal to epoch number
		epochNumber := sdk.BigEndianToUint64(key)
		// only return queued msgs within [beginEpoch, endEpoch]
		if epochNumber < beginEpoch || endEpoch < epochNumber {
			return false, nil
		}

		if accumulate {
			msgList := &types.QueuedMessageList{
				EpochNumber: epochNumber,
				Msgs:        k.GetEpochMsgs(ctx, epochNumber),
			}
			latestEpochMsgs = append(latestEpochMsgs, msgList)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &types.QueryLatestEpochMsgsResponse{
		LatestEpochMsgs: latestEpochMsgs,
		Pagination:      pageRes,
	}
	return resp, nil
}

// ValidatorLifecycle handles the QueryValidatorLifecycleRequest query
// TODO: test this API
func (k Keeper) ValidatorLifecycle(c context.Context, req *types.QueryValidatorLifecycleRequest) (*types.QueryValidatorLifecycleResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	valAddr, err := sdk.ValAddressFromBech32(req.ValAddr)
	if err != nil {
		return nil, err
	}
	lc := k.GetValLifecycle(ctx, valAddr)
	return &types.QueryValidatorLifecycleResponse{
		ValLife: lc,
	}, nil
}
