package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) FinalizedChainInfo(c context.Context, req *types.QueryFinalizedChainInfoRequest) (*types.QueryFinalizedChainInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// find the last finalised epoch
	finalizedEpoch, err := k.GetFinalizedEpoch(ctx)
	if err != nil {
		return nil, err
	}

	// find the chain info of this epoch
	chainInfo, err := k.GetEpochChainInfo(ctx, req.ChainId, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// TODO: construct inclusion proofs

	resp := &types.QueryFinalizedChainInfoResponse{
		FinalizedChainInfo: chainInfo,
	}
	return resp, nil
}
