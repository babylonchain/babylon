package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/x/epoching/types"
)

// CheckMsgCreateValidator performs checks on a given `MsgCreateValidator` message
// The checkpointing module will use this function to verify the `MsgCreateValidator` message
// inside a `MsgWrappedCreateValidator` message.
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.10/x/staking/keeper/msg_server.go#L34-L108)
func (k Keeper) CheckMsgCreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// ensure validator address is correctly encoded
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	// get parameters of the staking module
	sParams, err := k.stk.GetParams(ctx)
	if err != nil {
		return err
	}

	// check commission rate
	if msg.Commission.Rate.LT(sParams.MinCommissionRate) {
		return errorsmod.Wrapf(stakingtypes.ErrCommissionLTMinRate, "cannot set validator commission to less than minimum rate of %s", sParams.MinCommissionRate)
	}

	// ensure the validator operator was not registered before
	if _, err := k.stk.GetValidator(ctx, valAddr); err == nil {
		return stakingtypes.ErrValidatorOwnerExists
	}

	// check if the pubkey is correctly encoded
	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	// ensure the validator was not registered before
	if _, err := k.stk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); err == nil {
		return stakingtypes.ErrValidatorPubKeyExists
	}

	// ensure BondDemon is correct
	if msg.Value.Denom != sParams.BondDenom {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, sParams.BondDenom,
		)
	}

	// ensure description's length is valid
	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}

	// ensure public key type is supported
	cp := sdkCtx.ConsensusParams()
	if cp.Validator != nil {
		pkType := pk.Type()
		hasKeyType := false
		for _, keyType := range cp.Validator.PubKeyTypes {
			if pkType == keyType {
				hasKeyType = true
				break
			}
		}
		if !hasKeyType {
			return errorsmod.Wrapf(
				stakingtypes.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	// check validator
	validator, err := stakingtypes.NewValidator(valAddr.String(), pk, msg.Description)
	if err != nil {
		return err
	}

	// check if SetInitialCommission fails or not
	commission := stakingtypes.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, sdkCtx.BlockHeader().Time,
	)
	if _, err := validator.SetInitialCommission(commission); err != nil {
		return err
	}

	// sanity check on delegator address -- delegator is the same as validator
	delegatorAddr := sdk.AccAddress(valAddr)
	if err != nil {
		return err
	}

	balance := k.bk.GetBalance(ctx, delegatorAddr, msg.Value.GetDenom())
	if msg.Value.IsGTE(balance) {
		return types.ErrInsufficientBalance
	}

	return nil
}
