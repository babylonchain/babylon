package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/monitor module sentinel errors
var (
	ErrEpochNotFinishedYet = sdkerrors.Register(ModuleName, 1100, "Epoch not finished yet")
)
