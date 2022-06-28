package keeper

import (
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DropValidatorMsgDecorator defines an AnteHandler decorator that rejects all messages that might change the validator set.
type DropValidatorMsgDecorator struct{}

// NewDropValidatorMsgDecorator creates a new DropValidatorMsgDecorator
func NewDropValidatorMsgDecorator() *DropValidatorMsgDecorator {
	return &DropValidatorMsgDecorator{}
}

// AnteHandle performs an AnteHandler check that rejects all non-wrapped validator-related messages.
// It will reject the following types of messages:
// - MsgCreateValidator
// - MsgDelegate
// - MsgUndelegate
// - MsgBeginRedelegate
// TODO: after we bump to Cosmos SDK v0.46, add MsgCancelUnbondingDelegation
func (qmd DropValidatorMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		if qmd.IsValidatorRelatedMsg(msg) {
			return ctx, epochingtypes.ErrInvalidMsgType
		}
	}

	return next(ctx, tx, simulate)
}

// IsValidatorRelatedMsg checks if the given message is of non-wrapped type, which should be rejected
func (qmd DropValidatorMsgDecorator) IsValidatorRelatedMsg(msg sdk.Msg) bool {
	switch msg.(type) {
	case *stakingtypes.MsgCreateValidator, *stakingtypes.MsgDelegate, *stakingtypes.MsgUndelegate, *stakingtypes.MsgBeginRedelegate:
		return true
	default:
		return false
	}
}
