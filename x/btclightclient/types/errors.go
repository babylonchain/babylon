package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/btclightclient module sentinel errors
var (
	ErrHeaderDoesNotExist       = sdkerrors.Register(ModuleName, 1100, "header does not exist")
	ErrDuplicateHeader          = sdkerrors.Register(ModuleName, 1100, "duplicate header")
	ErrHeaderParentDoesNotExist = sdkerrors.Register(ModuleName, 1100, "header parent does not exist")
	ErrInvalidDifficulty        = sdkerrors.Register(ModuleName, 1100, "invalid difficulty bits")
)
