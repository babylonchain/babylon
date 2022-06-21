package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
func (qmd DropValidatorMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	panic("TODO: unimplemented")

	// return next(ctx, tx, simulate)
}
