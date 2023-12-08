package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
func (ms msgServer) WrappedDelegate(goCtx context.Context, msg *types.MsgWrappedDelegate) (*types.MsgWrappedDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// verification rules ported from staking module
	valAddr, valErr := sdk.ValAddressFromBech32(msg.Msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	if _, err := ms.stk.GetValidator(ctx, valAddr); err != nil {
		return nil, err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Msg.DelegatorAddress); err != nil {
		return nil, err
	}
	bondDenom, err := ms.stk.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}

	blockHeight := uint64(ctx.HeaderInfo().Height)
	if blockHeight == 0 {
		return nil, types.ErrZeroEpochMsg
	}
	blockTime := ctx.HeaderInfo().Time

	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(blockHeight, blockTime, txid, msg)
	if err != nil {
		return nil, err
	}

	ms.EnqueueMsg(ctx, queuedMsg)

	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedDelegate{
			DelegatorAddress: msg.Msg.DelegatorAddress,
			ValidatorAddress: msg.Msg.ValidatorAddress,
			Amount:           msg.Msg.Amount.Amount.Uint64(),
			Denom:            msg.Msg.Amount.GetDenom(),
			EpochBoundary:    ms.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedDelegateResponse{}, nil
}

// WrappedUndelegate handles the MsgWrappedUndelegate request
func (ms msgServer) WrappedUndelegate(goCtx context.Context, msg *types.MsgWrappedUndelegate) (*types.MsgWrappedUndelegateResponse, error) {
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
	if _, err := ms.stk.ValidateUnbondAmount(ctx, delegatorAddress, valAddr, msg.Msg.Amount.Amount); err != nil {
		return nil, err
	}
	bondDenom, err := ms.stk.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}

	blockHeight := uint64(ctx.HeaderInfo().Height)
	if blockHeight == 0 {
		return nil, types.ErrZeroEpochMsg
	}
	blockTime := ctx.HeaderInfo().Time

	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(blockHeight, blockTime, txid, msg)
	if err != nil {
		return nil, err
	}

	ms.EnqueueMsg(ctx, queuedMsg)

	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedUndelegate{
			DelegatorAddress: msg.Msg.DelegatorAddress,
			ValidatorAddress: msg.Msg.ValidatorAddress,
			Amount:           msg.Msg.Amount.Amount.Uint64(),
			Denom:            msg.Msg.Amount.GetDenom(),
			EpochBoundary:    ms.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedUndelegateResponse{}, nil
}

// WrappedBeginRedelegate handles the MsgWrappedBeginRedelegate request
func (ms msgServer) WrappedBeginRedelegate(goCtx context.Context, msg *types.MsgWrappedBeginRedelegate) (*types.MsgWrappedBeginRedelegateResponse, error) {
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
	if _, err := ms.stk.ValidateUnbondAmount(ctx, delegatorAddress, valSrcAddr, msg.Msg.Amount.Amount); err != nil {
		return nil, err
	}
	bondDenom, err := ms.stk.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}
	if _, err := sdk.ValAddressFromBech32(msg.Msg.ValidatorDstAddress); err != nil {
		return nil, err
	}

	blockHeight := uint64(ctx.HeaderInfo().Height)
	if blockHeight == 0 {
		return nil, types.ErrZeroEpochMsg
	}
	blockTime := ctx.HeaderInfo().Time

	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(blockHeight, blockTime, txid, msg)
	if err != nil {
		return nil, err
	}

	ms.EnqueueMsg(ctx, queuedMsg)
	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedBeginRedelegate{
			DelegatorAddress:            msg.Msg.DelegatorAddress,
			SourceValidatorAddress:      msg.Msg.ValidatorSrcAddress,
			DestinationValidatorAddress: msg.Msg.ValidatorDstAddress,
			Amount:                      msg.Msg.Amount.Amount.Uint64(),
			Denom:                       msg.Msg.Amount.GetDenom(),
			EpochBoundary:               ms.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedBeginRedelegateResponse{}, nil
}

// WrappedCancelUnbondingDelegation handles the MsgWrappedCancelUnbondingDelegation request
func (ms msgServer) WrappedCancelUnbondingDelegation(goCtx context.Context, msg *types.MsgWrappedCancelUnbondingDelegation) (*types.MsgWrappedCancelUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// verification rules ported from staking module
	if _, err := sdk.AccAddressFromBech32(msg.Msg.DelegatorAddress); err != nil {
		return nil, err
	}

	if _, err := sdk.ValAddressFromBech32(msg.Msg.ValidatorAddress); err != nil {
		return nil, err
	}

	if !msg.Msg.Amount.IsValid() || !msg.Msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid amount",
		)
	}

	if msg.Msg.CreationHeight <= 0 {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid height",
		)
	}

	bondDenom, err := ms.stk.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Msg.Amount.Denom, bondDenom,
		)
	}

	blockHeight := uint64(ctx.BlockHeader().Height)
	if blockHeight == 0 {
		return nil, types.ErrZeroEpochMsg
	}
	blockTime := ctx.BlockTime()
	txid := tmhash.Sum(ctx.TxBytes())
	queuedMsg, err := types.NewQueuedMessage(blockHeight, blockTime, txid, msg)
	if err != nil {
		return nil, err
	}

	ms.EnqueueMsg(ctx, queuedMsg)
	err = ctx.EventManager().EmitTypedEvents(
		&types.EventWrappedCancelUnbondingDelegation{
			DelegatorAddress: msg.Msg.DelegatorAddress,
			ValidatorAddress: msg.Msg.ValidatorAddress,
			Amount:           msg.Msg.Amount.Amount.Uint64(),
			CreationHeight:   msg.Msg.CreationHeight,
			EpochBoundary:    ms.GetEpoch(ctx).GetLastBlockHeight(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWrappedCancelUnbondingDelegationResponse{}, nil
}

// UpdateParams updates the params.
// TODO investigate when it is the best time to update the params. We can update them
// when the epoch changes, but we can also update them during the epoch and extend
// the epoch duration.
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
