package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/checkpointing module sentinel errors
var (
	ErrCkptDoesNotExist = sdkerrors.Register(ModuleName, 1201, "raw checkpoint does not exist")
	ErrCkptsDoNotExist  = sdkerrors.Register(ModuleName, 1202, "raw checkpoints do not exist")
)
