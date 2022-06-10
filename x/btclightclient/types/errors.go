package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/btclightclient module sentinel errors
var (
	ErrInvalidHeader = sdkerrors.Register(ModuleName, 1100, "invalid header")
)
