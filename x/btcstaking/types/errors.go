package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcstaking module sentinel errors
var (
	ErrBTCValNotFound         = errorsmod.Register(ModuleName, 1100, "the BTC validator is not found")
	ErrBTCDelNotFound         = errorsmod.Register(ModuleName, 1101, "the BTC delegation is not found")
	ErrDuplicatedBTCVal       = errorsmod.Register(ModuleName, 1102, "the BTC validator has already been registered")
	ErrBTCStakingNotActivated = errorsmod.Register(ModuleName, 1103, "the BTC staking protocol is not activated yet")
)
