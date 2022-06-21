package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WrappedAnteHandler is the wrapped AnteHandler that implements the `AnteDecorator` interface, which has a single function `AnteHandle`.
// It allows us to chain an existing AnteHandler with other decorators by using `sdk.ChainAnteDecorators`.
type WrappedAnteHandler struct {
	ah sdk.AnteHandler
}

// NewWrappedAnteHandler creates a new WrappedAnteHandler for a given AnteHandler.
func NewWrappedAnteHandler(ah sdk.AnteHandler) WrappedAnteHandler {
	return WrappedAnteHandler{ah}
}

func (wah WrappedAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	newCtx, err = wah.ah(ctx, tx, simulate)
	if err != nil {
		return newCtx, err
	}
	return next(newCtx, tx, simulate)
}
