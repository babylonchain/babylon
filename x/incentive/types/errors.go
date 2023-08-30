package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/incentive module sentinel errors
var (
	ErrSample = errorsmod.Register(ModuleName, 1100, "sample error")
)
