package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	// This should be a reference to Keeper
	k Keeper
}

func (m msgServer) InsertHeaders(ctx context.Context, msg *types.MsgInsertHeaders) (*types.MsgInsertHeadersResponse, error) {
	if msg == nil {
		return nil, types.ErrEmptyMessage.Wrapf("message is nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.InsertHeaders(sdkCtx, msg.Headers)

	if err != nil {
		return nil, err
	}
	return &types.MsgInsertHeadersResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}
