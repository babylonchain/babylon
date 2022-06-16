package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func noOpAnteDecorator() sdk.AnteHandler {
	return func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
}

func TestQueueMsgDecorator(t *testing.T) {
	panic("TODO: unimplemented")
}
