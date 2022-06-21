package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/checkpointing module sentinel errors
var (
	ErrBlsSigDoesNotExist       = sdkerrors.Register(ModuleName, 1100, "bls sig does not exist")
	ErrBlsSigsEpochDoesNotExist = sdkerrors.Register(ModuleName, 1101, "bls sig epoch does not exist")
)
