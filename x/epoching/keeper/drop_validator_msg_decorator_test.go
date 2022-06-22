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

func TestDropValidatorMsgDecorator(t *testing.T) {
<<<<<<< HEAD
	panic("TODO: unimplemented")
=======
	t.Errorf("TODO: unimplemented")
>>>>>>> main
}
