package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/checkpointing module sentinel errors
var (
	ErrBlsSigDoesNotExist       = sdkerrors.Register(ModuleName, 1200, "bls sig does not exist")
	ErrBlsSigsDoNotExist        = sdkerrors.Register(ModuleName, 1201, "bls sigs do not exist")
	ErrBlsSigsEpochDoesNotExist = sdkerrors.Register(ModuleName, 1202, "bls sig epoch does not exist")

	ErrCkptDoesNotExist        = sdkerrors.Register(ModuleName, 1203, "raw checkpoint does not exist")
	ErrCkptsDoNotExist         = sdkerrors.Register(ModuleName, 1204, "raw checkpoints do not exist")
	ErrCkptsEpochDoesNotExist  = sdkerrors.Register(ModuleName, 1205, "raw checkpoint epoch does not exist")
	ErrCkptsStatusDoesNotExist = sdkerrors.Register(ModuleName, 1206, "raw checkpoint status does not exist")

	ErrLastConfirmedEpochDoesNotExist = sdkerrors.Register(ModuleName, 1207, "last confirmed epoch does not exist")
)
