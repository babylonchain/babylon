package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the params
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}
	if err := req.Params.Validate(); err != nil {
		return nil, govtypes.ErrInvalidProposalMsg.Wrapf("invalid parameter: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// WithdrawReward withdraws the reward of a given stakeholder
func (ms msgServer) WithdrawReward(goCtx context.Context, req *types.MsgWithdrawReward) (*types.MsgWithdrawRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get stakeholder type and address
	sType, err := types.NewStakeHolderTypeFromString(req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// withdraw reward, i.e., send withdrawable reward to the stakeholder address and clear the reward gauge
	withdrawnCoins, err := ms.withdrawReward(ctx, sType, addr)
	if err != nil {
		return nil, err
	}

	// all good
	return &types.MsgWithdrawRewardResponse{
		Coins: withdrawnCoins,
	}, nil
}
