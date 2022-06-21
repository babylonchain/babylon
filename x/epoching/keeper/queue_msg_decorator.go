package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueueMsgDecorator defines an AnteHandler decorator that rejects all messages that might change the validator set.
type QueueMsgDecorator struct{}

// NewQueueMsgDecorator creates a new QueueMsgDecorator
func NewQueueMsgDecorator() *QueueMsgDecorator {
	return &QueueMsgDecorator{}
}

// AnteHandle performs an AnteHandler check that rejects all non-wrapped validator-related messages.
// It will reject the following types of messages:
// - MsgCreateValidator
// - MsgDelegate
// - MsgUndelegate
// - MsgBeginRedelegate
func (qmd QueueMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	panic("TODO: unimplemented")

	// return next(ctx, tx, simulate)
}
