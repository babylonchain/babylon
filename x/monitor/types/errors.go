package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/monitor module sentinel errors
var (
	ErrEpochNotEnded         = errorsmod.Register(ModuleName, 1100, "Epoch not ended yet")
	ErrCheckpointNotReported = errorsmod.Register(ModuleName, 1101, "Checkpoint not reported yet")
)
