package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RewardGauge(goCtx context.Context, req *types.QueryRewardGaugeRequest) (*types.QueryRewardGaugeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// try to cast types for fields in the request
	sType, err := types.NewStakeHolderTypeFromString(req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// find reward gauge
	rg, err := k.GetRewardGauge(ctx, sType, address)
	if err != nil {
		return nil, err
	}

	return &types.QueryRewardGaugeResponse{RewardGauge: rg}, nil
}
