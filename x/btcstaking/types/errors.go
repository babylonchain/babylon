package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcstaking module sentinel errors
var (
	ErrBTCValNotFound           = errorsmod.Register(ModuleName, 1100, "the BTC validator is not found")
	ErrBTCDelegatorNotFound     = errorsmod.Register(ModuleName, 1101, "the BTC delegator is not found")
	ErrBTCDelegationNotFound    = errorsmod.Register(ModuleName, 1102, "the BTC delegation is not found")
	ErrDuplicatedBTCVal         = errorsmod.Register(ModuleName, 1103, "the BTC validator has already been registered")
	ErrBTCValAlreadySlashed     = errorsmod.Register(ModuleName, 1104, "the BTC validator has already been slashed")
	ErrBTCStakingNotActivated   = errorsmod.Register(ModuleName, 1105, "the BTC staking protocol is not activated yet")
	ErrBTCHeightNotFound        = errorsmod.Register(ModuleName, 1106, "the BTC height is not found")
	ErrReusedStakingTx          = errorsmod.Register(ModuleName, 1107, "the BTC staking tx is already used")
	ErrInvalidCovenantPK        = errorsmod.Register(ModuleName, 1108, "the BTC staking tx specifies a wrong covenant PK")
	ErrInvalidStakingTx         = errorsmod.Register(ModuleName, 1109, "the BTC staking tx is not valid")
	ErrInvalidSlashingTx        = errorsmod.Register(ModuleName, 1110, "the BTC slashing tx is not valid")
	ErrDuplicatedCovenantSig    = errorsmod.Register(ModuleName, 1111, "the BTC delegation has already received this covenant signature")
	ErrInvalidCovenantSig       = errorsmod.Register(ModuleName, 1112, "the covenant signature is not valid")
	ErrCommissionLTMinRate      = errorsmod.Register(ModuleName, 1113, "commission cannot be less than min rate")
	ErrCommissionGTMaxRate      = errorsmod.Register(ModuleName, 1114, "commission cannot be more than one")
	ErrInvalidDelegationState   = errorsmod.Register(ModuleName, 1115, "Unexpected delegation state")
	ErrInvalidUnbondingTx       = errorsmod.Register(ModuleName, 1116, "the BTC unbonding tx is not valid")
	ErrRewardDistCacheNotFound  = errorsmod.Register(ModuleName, 1117, "the reward distribution cache is not found")
	ErrEmptyValidatorList       = errorsmod.Register(ModuleName, 1118, "the validator list is empty")
	ErrInvalidProofOfPossession = errorsmod.Register(ModuleName, 1119, "the proof of possession is not valid")
	ErrDuplicatedValidator      = errorsmod.Register(ModuleName, 1120, "the staking request contains duplicated validator public key")
)
