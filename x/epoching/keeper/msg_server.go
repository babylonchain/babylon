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
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(txid, msg)
	if err != nil {
		return nil, err
	}

	k.EnqueueMsg(ctx, queuedMsg)
	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedDelegate{
			ValidatorAddress: msg.Msg.ValidatorAddress,
			Amount:           msg.Msg.Amount.Amount.Uint64(),
			Denom:            msg.Msg.Amount.GetDenom(),
			EpochBoundary:    k.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedDelegateResponse{}, nil
}

// WrappedUndelegate handles the MsgWrappedUndelegate request
func (k msgServer) WrappedUndelegate(goCtx context.Context, msg *types.MsgWrappedUndelegate) (*types.MsgWrappedUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(txid, msg)
	if err != nil {
		return nil, err
	}

	k.EnqueueMsg(ctx, queuedMsg)
	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedUndelegate{
			ValidatorAddress: msg.Msg.ValidatorAddress,
			Amount:           msg.Msg.Amount.Amount.Uint64(),
			Denom:            msg.Msg.Amount.GetDenom(),
			EpochBoundary:    k.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedUndelegateResponse{}, nil
}

// WrappedBeginRedelegate handles the MsgWrappedBeginRedelegate request
func (k msgServer) WrappedBeginRedelegate(goCtx context.Context, msg *types.MsgWrappedBeginRedelegate) (*types.MsgWrappedBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(txid, msg)
	if err != nil {
		return nil, err
	}

	k.EnqueueMsg(ctx, queuedMsg)
	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedBeginRedelegate{
			SourceValidatorAddress:      msg.Msg.ValidatorSrcAddress,
			DestinationValidatorAddress: msg.Msg.ValidatorDstAddress,
			Amount:                      msg.Msg.Amount.Amount.Uint64(),
			Denom:                       msg.Msg.Amount.GetDenom(),
			EpochBoundary:               k.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedBeginRedelegateResponse{}, nil
}
