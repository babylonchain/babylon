package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	// This should be a reference to Keeper
	k Keeper
}

func (m msgServer) canInsertHeaders(sdkCtx sdk.Context, reporterAddress sdk.AccAddress) bool {
	params := m.k.GetParams(sdkCtx)

	if params.AllowAllReporters() {
		return true
	}

	var allowInsertHeaders bool = false
	for _, addr := range params.InsertHeadersAllowList {
		if sdk.MustAccAddressFromBech32(addr).Equals(reporterAddress) {
			allowInsertHeaders = true
		}
	}

	return allowInsertHeaders
}

func (m msgServer) InsertHeaders(ctx context.Context, msg *types.MsgInsertHeaders) (*types.MsgInsertHeadersResponse, error) {
	if msg == nil {
		return nil, types.ErrEmptyMessage.Wrapf("message is nil")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := msg.ValidateStateless(); err != nil {
		return nil, types.ErrInvalidMessageFormat.Wrapf("invalid insert header message: %v", err)
	}

	reporterAddress := msg.ReporterAddress()

	if !m.canInsertHeaders(sdkCtx, reporterAddress) {
		return nil, types.ErrUnauthorizedReporter.Wrapf("reporter %s is not authorized to insert headers", reporterAddress)
	}

	err := m.k.InsertHeaders(sdkCtx, msg.Headers)

	if err != nil {
		return nil, err
	}
	return &types.MsgInsertHeadersResponse{}, nil
}

func (ms msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.k.authority, req.Authority)
	}
	if err := req.Params.Validate(); err != nil {
		return nil, govtypes.ErrInvalidProposalMsg.Wrapf("invalid parameter: %v", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.k.SetParams(sdkCtx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}
