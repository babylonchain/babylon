package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/epoching/types"
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

// WrappedDelegate handles the MsgWrappedDelegate request
func (k msgServer) WrappedDelegate(goCtx context.Context, msg *types.MsgWrappedDelegate) (*types.MsgWrappedDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// wrapped -> unwrapped -> QueuedMessage
	queuedMsg := types.QueuedMessage{
		MsgId: msg.GetSignBytes(), // TODO: not sure if this can be a good MsgId or not
		Msg: &types.QueuedMessage_MsgDelegate{
			MsgDelegate: msg.Msg,
		},
	}

	// enqueue msg
	if err := k.EnqueueMsg(ctx, queuedMsg); err != nil {
		return nil, err
	}

	return &types.MsgWrappedDelegateResponse{}, nil
}

// WrappedUndelegate handles the MsgWrappedUndelegate request
func (k msgServer) WrappedUndelegate(goCtx context.Context, msg *types.MsgWrappedUndelegate) (*types.MsgWrappedUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// wrapped -> unwrapped -> QueuedMessage
	queuedMsg := types.QueuedMessage{
		MsgId: msg.GetSignBytes(), // TODO: not sure if this can be a good MsgId or not
		Msg: &types.QueuedMessage_MsgUndelegate{
			MsgUndelegate: msg.Msg,
		},
	}

	// enqueue msg
	if err := k.EnqueueMsg(ctx, queuedMsg); err != nil {
		return nil, err
	}

	return &types.MsgWrappedUndelegateResponse{}, nil
}

// WrappedBeginRedelegate handles the MsgWrappedBeginRedelegate request
func (k msgServer) WrappedBeginRedelegate(goCtx context.Context, msg *types.MsgWrappedBeginRedelegate) (*types.MsgWrappedBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// wrapped -> unwrapped -> QueuedMessage
	queuedMsg := types.QueuedMessage{
		MsgId: msg.GetSignBytes(), // TODO: not sure if this can be a good MsgId or not
		Msg: &types.QueuedMessage_MsgBeginRedelegate{
			MsgBeginRedelegate: msg.Msg,
		},
	}

	// enqueue msg
	if err := k.EnqueueMsg(ctx, queuedMsg); err != nil {
		return nil, err
	}

	return &types.MsgWrappedBeginRedelegateResponse{}, nil
}
