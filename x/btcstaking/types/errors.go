package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcstaking module sentinel errors
var (
	ErrBTCValNotFound         = errorsmod.Register(ModuleName, 1100, "the BTC validator is not found")
	ErrBTCDelNotFound         = errorsmod.Register(ModuleName, 1101, "the BTC delegation is not found")
	ErrDuplicatedBTCVal       = errorsmod.Register(ModuleName, 1102, "the BTC validator has already been registered")
	ErrBTCValAlreadySlashed   = errorsmod.Register(ModuleName, 1103, "the BTC validator has already been slashed")
	ErrBTCStakingNotActivated = errorsmod.Register(ModuleName, 1104, "the BTC staking protocol is not activated yet")
	ErrBTCHeightNotFound      = errorsmod.Register(ModuleName, 1105, "the BTC height is not found")
	ErrReusedStakingTx        = errorsmod.Register(ModuleName, 1106, "the BTC staking tx is already used")
	ErrInvalidJuryPK          = errorsmod.Register(ModuleName, 1107, "the BTC staking tx specifies a wrong jury PK")
	ErrInvalidStakingTx       = errorsmod.Register(ModuleName, 1108, "the BTC staking tx is not valid")
	ErrInvalidSlashingTx      = errorsmod.Register(ModuleName, 1109, "the BTC slashing tx is not valid")
	ErrDuplicatedJurySig      = errorsmod.Register(ModuleName, 1110, "the BTC delegation has already received jury signature")
	ErrInvalidJurySig         = errorsmod.Register(ModuleName, 1111, "the jury signature is not valid")
	ErrCommissionLTMinRate    = errorsmod.Register(ModuleName, 1112, "commission cannot be less than min rate")
	ErrInvalidDelegationState = errorsmod.Register(ModuleName, 1113, "Unexpected delegation state")
	ErrInvalidUnbodningTx     = errorsmod.Register(ModuleName, 1114, "the BTC unbonding tx is not valid")
)
