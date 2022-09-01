package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
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

	// verification rules ported from staking module
	valAddr, valErr := sdk.ValAddressFromBech32(msg.Msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	if _, found := k.stk.GetValidator(ctx, valAddr); !found {
		return nil, stakingtypes.ErrNoValidatorFound
	}
	if _, err := sdk.AccAddressFromBech32(msg.Msg.DelegatorAddress); err != nil {
		return nil, err
	}
	bondDenom := k.stk.BondDenom(ctx)
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}

	height := uint64(ctx.BlockHeight())
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(height, txid, msg)
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

	// verification rules ported from staking module
	valAddr, err := sdk.ValAddressFromBech32(msg.Msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.Msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	if _, err := k.stk.ValidateUnbondAmount(ctx, delegatorAddress, valAddr, msg.Msg.Amount.Amount); err != nil {
		return nil, err
	}
	bondDenom := k.stk.BondDenom(ctx)
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}

	height := uint64(ctx.BlockHeight())
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(height, txid, msg)
	if err != nil {
		return nil, err
	}

	k.EnqueueMsg(ctx, queuedMsg)
	k.UpdateValState(ctx, valAddr, types.ValStateUnbondingRequestSubmitted)

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

	// verification rules ported from staking module
	valSrcAddr, err := sdk.ValAddressFromBech32(msg.Msg.ValidatorSrcAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.Msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	if _, err := k.stk.ValidateUnbondAmount(ctx, delegatorAddress, valSrcAddr, msg.Msg.Amount.Amount); err != nil {
		return nil, err
	}
	bondDenom := k.stk.BondDenom(ctx)
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}
	if _, err := sdk.ValAddressFromBech32(msg.Msg.ValidatorDstAddress); err != nil {
		return nil, err
	}

	height := uint64(ctx.BlockHeight())
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(height, txid, msg)
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
