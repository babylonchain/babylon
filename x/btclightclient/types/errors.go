package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/btclightclient module sentinel errors
var (
	ErrHeaderDoesNotExist       = sdkerrors.Register(ModuleName, 1100, "header does not exist")
	ErrDuplicateHeader          = sdkerrors.Register(ModuleName, 1101, "duplicate header")
	ErrHeaderParentDoesNotExist = sdkerrors.Register(ModuleName, 1102, "header parent does not exist")
	ErrInvalidDifficulty        = sdkerrors.Register(ModuleName, 1103, "invalid difficulty bits")
	ErrEmptyMessage             = sdkerrors.Register(ModuleName, 1104, "empty message provided")
	ErrInvalidAncestor          = sdkerrors.Register(ModuleName, 1105, "invalid ancestor provided")
)
