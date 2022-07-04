package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/btccheckpoint module sentinel errors
var (
	ErrInvalidCheckpointProof = sdkerrors.Register(ModuleName, 1100, "Invalid checkpoint proof")
)
