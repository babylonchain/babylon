package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcstaking module sentinel errors
var (
	ErrBTCValNotFound                  = errorsmod.Register(ModuleName, 1100, "the BTC validator is not found")
	ErrBTCDelegatorNotFound            = errorsmod.Register(ModuleName, 1101, "the BTC delegator is not found")
	ErrBTCDelegationNotFound           = errorsmod.Register(ModuleName, 1102, "the BTC delegation is not found")
	ErrDuplicatedBTCVal                = errorsmod.Register(ModuleName, 1103, "the BTC validator has already been registered")
	ErrBTCValAlreadySlashed            = errorsmod.Register(ModuleName, 1104, "the BTC validator has already been slashed")
	ErrBTCStakingNotActivated          = errorsmod.Register(ModuleName, 1105, "the BTC staking protocol is not activated yet")
	ErrBTCHeightNotFound               = errorsmod.Register(ModuleName, 1106, "the BTC height is not found")
	ErrReusedStakingTx                 = errorsmod.Register(ModuleName, 1107, "the BTC staking tx is already used")
	ErrInvalidJuryPK                   = errorsmod.Register(ModuleName, 1108, "the BTC staking tx specifies a wrong jury PK")
	ErrInvalidStakingTx                = errorsmod.Register(ModuleName, 1109, "the BTC staking tx is not valid")
	ErrInvalidSlashingTx               = errorsmod.Register(ModuleName, 1110, "the BTC slashing tx is not valid")
	ErrDuplicatedJurySig               = errorsmod.Register(ModuleName, 1111, "the BTC delegation has already received jury signature")
	ErrInvalidJurySig                  = errorsmod.Register(ModuleName, 1112, "the jury signature is not valid")
	ErrCommissionLTMinRate             = errorsmod.Register(ModuleName, 1113, "commission cannot be less than min rate")
	ErrInvalidDelegationState          = errorsmod.Register(ModuleName, 1114, "Unexpected delegation state")
	ErrInvalidUnbodningTx              = errorsmod.Register(ModuleName, 1115, "the BTC unbonding tx is not valid")
	ErrUnbondingDuplicatedValidatorSig = errorsmod.Register(ModuleName, 1116, "the BTC undelegation has already received validator signature")
	ErrUnbodningInvalidValidatorSig    = errorsmod.Register(ModuleName, 1117, "the validator signature is not valid")
	ErrUnbondingUnexpectedValidatorSig = errorsmod.Register(ModuleName, 1118, "the BTC undelegation did not receive validator signature yet")
	ErrRewardDistCacheNotFound         = errorsmod.Register(ModuleName, 1119, "the reward distribution cache is not found")
)
