package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/checkpointing module sentinel errors
var (
	ErrCkptDoesNotExist   = sdkerrors.Register(ModuleName, 1201, "raw checkpoint does not exist")
	ErrCkptAlreadyExist   = sdkerrors.Register(ModuleName, 1202, "raw checkpoint already exists")
	ErrCkptHashNotEqual   = sdkerrors.Register(ModuleName, 1203, "hash does not equal to raw checkpoint")
	ErrBlsKeyDoesNotExist = sdkerrors.Register(ModuleName, 1204, "BLS public key does not exist")
	ErrBlsKeyAlreadyExist = sdkerrors.Register(ModuleName, 1205, "BLS public key already exists")
)
