package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RewardGauges(goCtx context.Context, req *types.QueryRewardGaugesRequest) (*types.QueryRewardGaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// try to cast address
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rgMap := map[string]*types.RewardGauge{}

	// find reward gauge
	for _, sType := range types.GetAllStakeholderTypes() {
		if !k.HasRewardGauge(ctx, sType, address) {
			continue
		}
		rg, err := k.GetRewardGauge(ctx, sType, address)
		if err != nil {
			// only programming error is possible
			panic("failed to get an existing reward gauge")
		}
		rgMap[sType.String()] = rg
	}

	return &types.QueryRewardGaugesResponse{RewardGauges: rgMap}, nil
}

func (k Keeper) BTCStakingGauge(goCtx context.Context, req *types.QueryBTCStakingGaugeRequest) (*types.QueryBTCStakingGaugeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// find gauge
	gauge, err := k.GetBTCStakingGauge(ctx, req.Height)
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCStakingGaugeResponse{Gauge: gauge}, nil
}

func (k Keeper) BTCTimestampingGauge(goCtx context.Context, req *types.QueryBTCTimestampingGaugeRequest) (*types.QueryBTCTimestampingGaugeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// find gauge
	gauge, err := k.GetBTCTimestampingGauge(ctx, req.EpochNum)
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCTimestampingGaugeResponse{Gauge: gauge}, nil
}
