package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btclightclient module sentinel errors
var (
	ErrHeaderDoesNotExist       = errorsmod.Register(ModuleName, 1100, "header does not exist")
	ErrHeaderParentDoesNotExist = errorsmod.Register(ModuleName, 1101, "parent for provided hash is not maintained")
	ErrEmptyMessage             = errorsmod.Register(ModuleName, 1102, "empty message provided")
	ErrInvalidProofOfWOrk       = errorsmod.Register(ModuleName, 1103, "provided header has invalid proof of work")
	ErrInvalidHeader            = errorsmod.Register(ModuleName, 1104, "provided header does not satisfy header validation rules")
	ErrChainWithNotEnoughWork   = errorsmod.Register(ModuleName, 1105, "provided chain has not enough work")
)
