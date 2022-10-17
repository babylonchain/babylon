package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/btclightclient module sentinel errors
var (
	ErrHeaderDoesNotExist       = sdkerrors.Register(ModuleName, 1100, "header does not exist")
	ErrDuplicateHeader          = sdkerrors.Register(ModuleName, 1101, "header with provided hash already exists")
	ErrHeaderParentDoesNotExist = sdkerrors.Register(ModuleName, 1102, "parent for provided hash is not maintained")
	ErrInvalidDifficulty        = sdkerrors.Register(ModuleName, 1103, "invalid difficulty bits")
	ErrEmptyMessage             = sdkerrors.Register(ModuleName, 1104, "empty message provided")
	ErrInvalidProofOfWOrk       = sdkerrors.Register(ModuleName, 1105, "provided header has invalid proof of work")
)
