package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/monitor module sentinel errors
var (
	ErrEpochNotEnded         = sdkerrors.Register(ModuleName, 1100, "Epoch not ended yet")
	ErrCheckpointNotReported = sdkerrors.Register(ModuleName, 1101, "Checkpoint not reported yet")
)
