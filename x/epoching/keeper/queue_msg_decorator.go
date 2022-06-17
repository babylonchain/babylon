package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueueMsgDecorator defines an AnteHandler decorator for queueing messages that might change the validator set.
type QueueMsgDecorator struct {
	epochingKeeper Keeper
}

func NewQueueMsgDecorator(ek Keeper) *QueueMsgDecorator {
	return &QueueMsgDecorator{
		epochingKeeper: ek,
	}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message that is blocked.
// Right now, we block MsgTimeoutOnClose due to incorrect behavior that could occur if a packet is re-enabled.
func (qmd QueueMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	panic("TODO: unimplemented")

	// return next(ctx, tx, simulate)
}
