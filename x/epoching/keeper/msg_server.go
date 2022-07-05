package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

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

	// get msg in bytes
	msgBytes, err := k.cdc.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// wrapped -> unwrapped -> QueuedMessage
	queuedMsg := types.QueuedMessage{
		TxId:  tmhash.Sum(ctx.TxBytes()),
		MsgId: tmhash.Sum(msgBytes),
		Msg: &types.QueuedMessage_MsgDelegate{
			MsgDelegate: msg.Msg,
		},
	}

	// enqueue msg
	k.EnqueueMsg(ctx, queuedMsg)

	return &types.MsgWrappedDelegateResponse{}, nil
}

// WrappedUndelegate handles the MsgWrappedUndelegate request
func (k msgServer) WrappedUndelegate(goCtx context.Context, msg *types.MsgWrappedUndelegate) (*types.MsgWrappedUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get msg in bytes
	msgBytes, err := k.cdc.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// wrapped -> unwrapped -> QueuedMessage
	queuedMsg := types.QueuedMessage{
		TxId:  tmhash.Sum(ctx.TxBytes()),
		MsgId: tmhash.Sum(msgBytes),
		Msg: &types.QueuedMessage_MsgUndelegate{
			MsgUndelegate: msg.Msg,
		},
	}

	// enqueue msg
	k.EnqueueMsg(ctx, queuedMsg)

	return &types.MsgWrappedUndelegateResponse{}, nil
}

// WrappedBeginRedelegate handles the MsgWrappedBeginRedelegate request
func (k msgServer) WrappedBeginRedelegate(goCtx context.Context, msg *types.MsgWrappedBeginRedelegate) (*types.MsgWrappedBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get msg in bytes
	msgBytes, err := k.cdc.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// wrapped -> unwrapped -> QueuedMessage
	queuedMsg := types.QueuedMessage{
		TxId:  tmhash.Sum(ctx.TxBytes()),
		MsgId: tmhash.Sum(msgBytes),
		Msg: &types.QueuedMessage_MsgBeginRedelegate{
			MsgBeginRedelegate: msg.Msg,
		},
	}

	// enqueue msg
	k.EnqueueMsg(ctx, queuedMsg)

	return &types.MsgWrappedBeginRedelegateResponse{}, nil
}
