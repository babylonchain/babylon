package btccheckpoint

import (
	"fmt"

	"github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler ...
func NewHandler(k keeper.Keeper) sdk.Handler {

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {

		switch msg := msg.(type) {
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
