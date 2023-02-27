package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btclightclient module sentinel errors
var (
	ErrHeaderDoesNotExist       = errorsmod.Register(ModuleName, 1100, "header does not exist")
	ErrDuplicateHeader          = errorsmod.Register(ModuleName, 1101, "header with provided hash already exists")
	ErrHeaderParentDoesNotExist = errorsmod.Register(ModuleName, 1102, "parent for provided hash is not maintained")
	ErrInvalidDifficulty        = errorsmod.Register(ModuleName, 1103, "invalid difficulty bits")
	ErrEmptyMessage             = errorsmod.Register(ModuleName, 1104, "empty message provided")
	ErrInvalidProofOfWOrk       = errorsmod.Register(ModuleName, 1105, "provided header has invalid proof of work")
)
